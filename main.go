package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

// CH1 L04 https://www.boot.dev/lessons/861ada77-c583-42c8-a265-657f2c453103
// CH1 L05 https://www.boot.dev/lessons/8cf7315a-ffc0-4ce0-b482-5972ab695131
// CH1 L08 https://www.boot.dev/lessons/20709716-4d7c-47fe-b182-9bccf8436ddc
// CH1 L11 https://www.boot.dev/lessons/174d13f0-f887-46c6-a633-d963662fde39
// CH2 L01 https://www.boot.dev/lessons/a13ffa72-18b9-49f7-81a9-c5a17d007b3a
// CH3 L04 https://www.boot.dev/lessons/892b38f7-d154-4591-ac63-a9fbc2a38187
// CH4 L02 https://www.boot.dev/lessons/374ef0f7-1d2d-40b8-8cef-14e9ffd033ab
// CH4 L06 https://www.boot.dev/lessons/7cde3fa8-f38a-444e-92a6-83166a905cb0

// CH2 L01
type apiConfig struct {
	fileserverHits atomic.Int32
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

// CH4 L02
func handerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type cleanMessage struct {
		Body string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
	}

	// params is a struct with data populated successfully

	// No hi ha hagut errors.
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

	// CH4 L02
	// El chirp es valid
	payload := cleanMessage{
		Body: newText,
	}
	respondWithJSON(w, 200, payload)
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
	cfg.fileserverHits.Store(0)
}

func main() {
	apiCfg := apiConfig{}

	// CH1 L4-L5
	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileserverHandle()))
	mux.Handle("/assets", http.FileServer(http.Dir("./assets"))) // CH1 L05
	mux.HandleFunc("GET /api/healthz", handlerHealth)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)     // CH2 L1
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)        // CH2 L1
	mux.HandleFunc("POST /api/validate_chirp", handerValidateChirp) // CH4 L2

	// Create a new http.Server struct.
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	fmt.Printf("Server started on %v\n", server.Addr)

	// Use the server's ListenAndServe method to start the server
	log.Fatal(server.ListenAndServe())

}
