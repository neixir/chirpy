
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/neixir/chirpy/internal/auth"
	"github.com/neixir/chirpy/internal/database"
)

// CH1 L04 https://www.boot.dev/lessons/861ada77-c583-42c8-a265-657f2c453103
// CH1 L05 https://www.boot.dev/lessons/8cf7315a-ffc0-4ce0-b482-5972ab695131
// CH1 L08 https://www.boot.dev/lessons/20709716-4d7c-47fe-b182-9bccf8436ddc
// CH1 L11 https://www.boot.dev/lessons/174d13f0-f887-46c6-a633-d963662fde39
// CH2 L01 https://www.boot.dev/lessons/a13ffa72-18b9-49f7-81a9-c5a17d007b3a
// CH3 L04 https://www.boot.dev/lessons/892b38f7-d154-4591-ac63-a9fbc2a38187
// CH4 L02 https://www.boot.dev/lessons/374ef0f7-1d2d-40b8-8cef-14e9ffd033ab
// CH4 L06 https://www.boot.dev/lessons/7cde3fa8-f38a-444e-92a6-83166a905cb0
// CH5 L01 i seguents per PostgreSQL, Goose, SLQC
// CH5 L09 https://www.boot.dev/lessons/341b80d4-556f-4c5b-8afc-ffd12d5238c2
// CH5 L10 https://www.boot.dev/lessons/0a07a4a3-c11f-429f-ac70-52fa2e016bc0
// CH6 L07 https://www.boot.dev/lessons/0689e0d0-bdb1-4cc8-b577-f0dd0535ad00
// CH6 L12 https://www.boot.dev/lessons/f7285cef-5185-4b15-b5fc-9533ccaafe8a
// CH7 L01 https://www.boot.dev/lessons/be14c814-e6c2-4b96-a361-e33bcfe71f00
// CH7 L04 https://www.boot.dev/lessons/61628ee7-a227-45a2-ab79-2721a52db32a
// CH8 L10 https://www.boot.dev/lessons/1304e939-bf50-48d3-a351-b35faafc267d

// CH2 L01
type apiConfig struct {
	fileserverHits atomic.Int32			// ja no cal?
	db             *database.Queries
	user			database.User		// s'haura de treure?
	secret			string
}

// TODO renombrar a userResponse
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token		string	`json:"token"`
	RefreshToken	string	`json:"refresh_token"`
	IsChirpyRed bool	`json:"is_chirpy_red"`
}

// TODO renombrar (i potser no es fa servir)
type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body     string    `json:"body"`
	UserID        uuid.UUID `json:"user_id"`
}

type polkaWebhook struct {
	Event	string	`json:"event"`
	Data	polkaWebhookData `json:"data"`
}

type polkaWebhookData struct {
	UserID        uuid.UUID `json:"user_id"`
}

const DefaultExpiresInSeconds = 60 * 60		// 1h
const MaxExpiresInSeconds = 60 * 60

// CH2 L01
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// CH1 L11
func handlerHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	w.WriteHeader(200)

	//fmt.Fprintf(w, "OK")
	w.Write([]byte("OK"))

}

func cleanBody(text string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Split(text, " ")
	for i, word := range words {
		for _, profanity := range profaneWords {
			if strings.ToLower(word) == profanity {
				words[i] = "****"
			}
		}
	}

	return strings.Join(words, " ") //newText
}

// CH5 L05
func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	// Aixo  de parameters/decoder/decode es repeteix molt, fer-ne metode
	type parameters struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong decoding parameters")
		return
	}
	
	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "Something went wrong hashing password")
		return
	}

	args := database.CreateUserParams{
		Email: params.Email,
		HashedPassword: hashed_password,
	}

	user, err := cfg.db.CreateUser(r.Context(), args)
	if err != nil {
		fmt.Println(err.Error())
		respondWithError(w, 500, "Something went wrong creating user")
		return
	}

	// Creem "payload" a partir de "user".
	// Es el mateix, pero sense password
	payload := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
	}
	
	// cfg.user = user

	respondWithJSON(w, 201, payload)

}

// CH5 L06
func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
		// UserID uuid.UUID `json:"user_id"`	// segurament no es fara servir
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong decoding parameters")
		return
	}

	// CH6 L07
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		// Responem amb l'error exacte, pero en prod hauriem de loguejar l'error i retornar ""
		respondWithError(w, 500, err.Error())
		return
	}
	
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		// Responem amb l'error exacte, pero en prod hauriem de loguejar l'error i retornar ""
		respondWithError(w, 401, err.Error())
		return
	}

	// CH4 L02
	// No hi ha hagut errors obtenint el cos del request
	// Aqui ara desariem params.Body a la base de dades o el que sigui,
	// o continuem comprovant si el chirp es valid
	// For example, if the Chirp is too long, respond with a 400 code and this body:
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	// CH4 L06
	// Assuming the length validation passed, replace any of the following words in the Chirp with the static 4-character string ****:
	newText := cleanBody(params.Body)

	// CH4 L02 El chirp es valid
	// CH5 L06 If the chirp is valid, you should save it in the database
	// Suposem que acabem de crear un usuari, si no fallara
	// (hauriem de posar-ho un middleware, pero com que al curs no ho diu, no ho fem)
	args := database.CreateChirpParams{
		Body: newText,
		UserID: userID,	//cfg.user.ID,	//uuid.New(),
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), args)
	if err != nil {
		respondWithError(w, 500, "Something went wrong saving chirp")
		return
	}

	respondWithJSON(w, 201, chirp)
}

// CH5 L09
func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "Something went wrong retrieving chirps")
		return
	}

	respondWithJSONArray(w, 200, chirps)
}

// CH5 L10
func (cfg *apiConfig) handlerGetOneChirp(w http.ResponseWriter, r *http.Request) {
	chirp_id_string := r.PathValue("chirpID")
	// sera "" si no troba "chirpID" -- retornar 500 pq es error del servidor

	// https://stackoverflow.com/a/62952994
	chirp_id, _ := uuid.Parse(chirp_id_string)
	// posar err si volem mirar que el format sigui correcte -- igualment seria 500 error intern

	chirp, err := cfg.db.GetOneChirp(r.Context(), chirp_id)
	if err != nil {
		//respondWithError(w, 500, "Something went wrong retrieving chirp")
		w.WriteHeader(404)
		return
	}

	respondWithJSON(w, 200, chirp)
}

// This endpoint should allow a user to login.
// In a future exercise, this endpoint will be used to give the user a token that
// they can use to make authenticated requests.
// For now, let's just make sure password validation is working.
func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
    type parameters struct {
        Email string `json:"email"`
        Password string `json:"password"`
    }

    decoder := json.NewDecoder(r.Body)
    params := parameters{}
    err := decoder.Decode(&params)
    if err != nil {
        respondWithError(w, 500, "Something went wrong decoding parameters")
        return
    }

	wrongEmailOrPassword := false
	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		wrongEmailOrPassword = true
	}
    
    // Once you have the user, check to see if their password matches the stored hash using your internal package.
    match := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	
    // If either the user lookup or the password comparison errors, just return a 401 Unauthorized response
    // with the message "Incorrect email or password".
    if match != nil {
		wrongEmailOrPassword = true
	}

	if wrongEmailOrPassword {
		respondWithError(w, 401, "Incorrect email or password")
		return
    } else { 
		token, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(DefaultExpiresInSeconds) * time.Second)
		if err != nil {
			respondWithError(w, 500, "Error creating token")
			return
		}

		// Refresh token
		refreshToken, _ := auth.MakeRefreshToken()
		
		args := database.CreateRefreshTokenParams{
			Token: refreshToken,
			UserID: user.ID,
			ExpiresAt: time.Now().UTC().AddDate(0, 0, 60),
		}

		_, err = cfg.db.CreateRefreshToken(r.Context(), args)
		if err != nil {
			respondWithError(w, 500, "Error saving refresh token")
			return
		}

		// Es el mateix struct que User, pero sense password i amb token
        payload := User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt, //Format(time.RFC3339),
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			Token:		token,
			RefreshToken:	refreshToken,
			IsChirpyRed: user.IsChirpyRed.Bool,
        }

        respondWithJSON(w, 200, payload)
    }
}

// CH6 L12
func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Tambe diu "does not accept a request body",
	// potser podriem veure si hi ha body i en aquest cas sortir sense fer res

	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		// Responem amb l'error exacte, pero en prod hauriem de loguejar l'error i retornar ""
		respondWithError(w, 500, err.Error())
		return
	}

	doesNotExistOrExpired := false
	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), bearer)
	if err != nil {
		doesNotExistOrExpired = true
	}
	
	// fmt.Printf("now: %v\nexp: %v", time.Now().UTC(), refreshToken.ExpiresAt)
	
	if time.Now().UTC().After(refreshToken.ExpiresAt) {
		doesNotExistOrExpired = true
	}
	
	// Aixo no ho diu, pero si no ho mirem, "/api/revoke" no te cap efecte
	if refreshToken.RevokedAt.Valid {
		doesNotExistOrExpired = true
	}

	// If it doesn't exist, or if it's expired, respond with a 401 status code. 
	if doesNotExistOrExpired {
		w.WriteHeader(401)
		return
	}

	// Creem token now
	newJWT, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, time.Duration(DefaultExpiresInSeconds) * time.Second)
	if err != nil {
		respondWithError(w, 500, "Error creating token")
		return
	}

	// Otherwise, respond with a 200 code
	type refToken struct {
		Token string `json:"token"`
	}
	respondWithJSON(w, 200, refToken{Token:newJWT})

}

// CH6 L12
func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	// Tambe diu "does not accept a request body",
	// potser podriem veure si hi ha body i en aquest cas sortir sense fer res

	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		// Responem amb l'error exacte, pero en prod hauriem de loguejar l'error i retornar ""
		respondWithError(w, 500, err.Error())
		return
	}

	err = cfg.db.RevokeToken(r.Context(), bearer)
	if err != nil {
		// Responem amb l'error exacte, pero en prod hauriem de loguejar l'error i retornar ""
		respondWithError(w, 500, err.Error())
		return
	}

	// A 204 status means the request was successful but no body is returned.
	w.WriteHeader(204)
}

// CH7 L1
// Adaptat de handlerCreateUser
func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		// TODO Log error
		fmt.Println(err.Error())
		w.WriteHeader(401)
		return
	}

	userID, err := auth.ValidateJWT(bearer, cfg.secret)
	if err != nil {
		// TODO Log er.Error()
		w.WriteHeader(401)
		return
	}

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong decoding parameters")
		return
	}

	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "Something went wrong hashing password")
		return
	}

	args := database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: hashed_password,
		ID:             userID,
	}

	err = cfg.db.UpdateUser(r.Context(), args)
	if err != nil {
		fmt.Println(err.Error())
		respondWithError(w, 500, "Something went wrong updating user")
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 500, "Something went wrong searching for updated email")
		return
	}

	// Creem "payload" a partir de "user".
	// Es el mateix, pero sense password
	payload := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
	}

	respondWithJSON(w, 200, payload)
}

// CH7 L4
func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	bearer, err := auth.GetBearerToken(r.Header)
	if err != nil {
		// TODO Log error
		fmt.Println(err.Error())
		w.WriteHeader(401)
		return
	}

	userID, err := auth.ValidateJWT(bearer, cfg.secret)
	if err != nil {
		// TODO Log er.Error()
		w.WriteHeader(401)
		return
	}

	// Copiat o adaptat de hanldeGetOneChirp, potser fer-ne metode
	chirp_id_string := r.PathValue("chirpID")
	chirp_id, _ := uuid.Parse(chirp_id_string)
	chirp, err := cfg.db.GetOneChirp(r.Context(), chirp_id)
	if err != nil {
		// If the chirp is not found, return a 404 status code.
		// respondWithError(w, 500, "Something went wrong retrieving chirp")
		w.WriteHeader(404)
		return
	}

	if userID == chirp.UserID {
		err = cfg.db.DeleteChirp(r.Context(), chirp_id)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		// If the chirp is deleted successfully, return a 204 status code.
		w.WriteHeader(204)
		

	} else {
		// No es l'usuari
		w.WriteHeader(403)
		return
	}
}

func (cfg *apiConfig) handlerPolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := polkaWebhook{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong decoding parameters")
		return
	}

	// If the event is anything other than user.upgraded, the endpoint should immediately respond
	// with a 204 status code - we don't care about any other events.
	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	
	// TODO Comprovar que realment params.Data.UserID existeixi i sigui un UUID
	
	// If the event is user.upgraded, then it should update the user in the database, and mark that they are a Chirpy Red member.
	err = cfg.db.UpgradeUserToChirpyRed(r.Context(), params.Data.UserID)
	if err != nil {
		// If the user can't be found, the endpoint should respond with a 404 status code.
		// Pero l'error pot ser un altre aqui...
		fmt.Println(err.Error())
		w.WriteHeader(404)
		return
	}

	// If the user is upgraded successfully, the endpoint should respond with a 204 status code and an empty response body.
	w.WriteHeader(204)


	// Polka uses the response code to know whether or not the webhook was received successfully.
	// If the response code is anything other than 2XX, they'll retry the request

}

func fileserverHandle() http.Handler {
	// http://localhost:8080/app -> "./"
	return http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
}

// ... + CH3 L04
func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html :=
		`
<html>
	<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
	</body>
</html>
`
	fmt.Fprintf(w, html, cfg.fileserverHits.Load())

}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	platform := os.Getenv("PLATFORM")
	if platform != "dev" {
		respondWithError(w, 403, "aquest endpoint no funciona en aquest entorn")
		w.WriteHeader(403)
	}

	cfg.fileserverHits.Store(0)
	cfg.db.DeleteAllUsers(r.Context())
}

// CH4 L02 / L06
func respondWithError(w http.ResponseWriter, code int, message string) {
	type errorMessage struct {
		Error string `json:"error"`
	}

	respBody := errorMessage{
		Error: message,
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

// CH4 L06
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

// CH5 L09
func respondWithJSONArray[T any](w http.ResponseWriter, code int, payload []T) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON array: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func main() {
	// Llegim variables de .env
	// TODO Que doni error si el fitxer no existeix
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Errorf("Connectant al servidor -- %v", err)
		os.Exit(1)
	}

	dbQueries := database.New(db)
	apiCfg := apiConfig{
		db: dbQueries,
	}

	// CH6 L07
	apiCfg.secret = os.Getenv("SECRET")

	// CH1 L4-L5
	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileserverHandle()))
	mux.Handle("/assets", http.FileServer(http.Dir("./assets"))) // CH1 L05
	mux.HandleFunc("GET /api/healthz", handlerHealth)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)               // CH2 L01
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)                  // CH2 L01
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)               // CH5 L05 + CH6 L01
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)             // CH5 L06
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetAllChirps)             // CH5 L09
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetOneChirp)    // CH5 L10
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)                    // CH6 L01, L07, L12
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefreshToken)           // CH6 L12
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevokeToken)             // CH6 L12
	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)                // CH7 L01
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp) // CH7 L04
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerPolkaWebhooks) // CH8 L01

	// Create a new http.Server struct.
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	fmt.Printf("Server started on %v\n", server.Addr)

	// Use the server's ListenAndServe method to start the server
	log.Fatal(server.ListenAndServe())

}
