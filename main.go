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

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

// CH2 L01
type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	user			database.User
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body     string    `json:"body"`
	UserID        uuid.UUID `json:"user_id"`
}


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
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong decoding parameters")
		return
	}
	
	user, err := cfg.db.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 500, "Something went wrong creating user")
		return
	}

	// Creem "payload" a partir de "user".
	// En aquest cas es el mateix, pero no sempre sera aixi.
	payload := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	
	cfg.user = user

	respondWithJSON(w, 201, payload)

}

// CH5 L06
func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong decoding parameters")
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
		UserID: cfg.user.ID,	//uuid.New(),
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), args)
	if err != nil {
		respondWithError(w, 500, "Something went wrong saving chirp")
		return
	}

	payload := Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}
	respondWithJSON(w, 201, payload)
}

func fileserverHandle() http.Handler {
	// http://localhost:8080/app -> "./"
	return http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
}

// ... + CH3 L04
func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	//fmt.Fprintf(w, "Hits: %d", cfg.fileserverHits.Load())
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
		//respondWithError(w, 403, )
		fmt.Println("aquest endpoint no funciona en aquest entorn")
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

	// CH1 L4-L5
	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileserverHandle()))
	mux.Handle("/assets", http.FileServer(http.Dir("./assets"))) // CH1 L05
	mux.HandleFunc("GET /api/healthz", handlerHealth)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)     // CH2 L01
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)        // CH2 L01
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)     // CH5 L05

	// Create a new http.Server struct.
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	fmt.Printf("Server started on %v\n", server.Addr)

	// Use the server's ListenAndServe method to start the server
	log.Fatal(server.ListenAndServe())

}
