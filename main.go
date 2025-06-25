package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

// CH1 L04 https://www.boot.dev/lessons/861ada77-c583-42c8-a265-657f2c453103
// CH1 L05 https://www.boot.dev/lessons/8cf7315a-ffc0-4ce0-b482-5972ab695131
// CH1 L08 https://www.boot.dev/lessons/20709716-4d7c-47fe-b182-9bccf8436ddc
// CH1 L11 https://www.boot.dev/lessons/174d13f0-f887-46c6-a633-d963662fde39
// CH2 L01 https://www.boot.dev/lessons/a13ffa72-18b9-49f7-81a9-c5a17d007b3a
// CH3 L04 https://www.boot.dev/lessons/892b38f7-d154-4591-ac63-a9fbc2a38187
// CH4 L02 https://www.boot.dev/lessons/374ef0f7-1d2d-40b8-8cef-14e9ffd033ab

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

	type errorMessage struct {
		Error string `json:"error"`
	}

	type validMessage struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		// If any errors occur, it should respond with an appropriate HTTP status code and a JSON body of this shape:
		// { "error": "Something went wrong" }
		// log.Printf("Error decoding parameters: %s", err)
		respBody := errorMessage{
			Error: "Something went wrong",
		}

		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(dat)
		return
	}

	// params is a struct with data populated successfully

	// No hi ha hagut errors.
	// Aqui ara desariem params.Body a la base de dades o el que sigui,
	// o continuem comprovant si el chirp es valid
	// For example, if the Chirp is too long, respond with a 400 code and this body:
	if len(params.Body) > 140 {
		// Aixo esta basicament copiat de dalt,
		// es pot fer metode que li passem un missatge d'error i un status code i l'envii
		respBody := errorMessage{
			Error: "Chirp is too long",
		}

		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(dat)
		return
	}

	// El chirp es valid
	respBody := validMessage{
		Valid: true,
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func fileserverHandle() http.Handler {
	// http: //localhost:8080/app -> ""./""
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
