package main

import (
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
	// Write the Content-Type header
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Write the status code using w.WriteHeader
	w.WriteHeader(200)

	// Write the body text using w.Write
	//fmt.Fprintf(w, "OK")
	w.Write([]byte("OK"))

}

func fileserverHandle() http.Handler {
	// http: //localhost:8080/app -> ""./""
	return http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hits: %d", cfg.fileserverHits.Load())
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
	mux.HandleFunc("GET /api/metrics", apiCfg.metricsHandler) // CH2 L1
	mux.HandleFunc("POST /api/reset", apiCfg.resetHandler)    // CH2 L1

	// Create a new http.Server struct.
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	// Use the server's ListenAndServe method to start the server
	log.Fatal(server.ListenAndServe())

}
