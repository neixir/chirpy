package main

import (
	"log"
	"net/http"
)

// CH1 L04 https://www.boot.dev/lessons/861ada77-c583-42c8-a265-657f2c453103
// CH1 L05 https://www.boot.dev/lessons/8cf7315a-ffc0-4ce0-b482-5972ab695131
// CH1 L08 https://www.boot.dev/lessons/20709716-4d7c-47fe-b182-9bccf8436ddc
// CH1 L11 https://www.boot.dev/lessons/174d13f0-f887-46c6-a633-d963662fde39

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

func main() {
	// CH1 L4-L5
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handlerHealth)

	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir(".")))) // CH1 L05+L11
	mux.Handle("/assets", http.FileServer(http.Dir("./assets")))                   // CH1 L05

	// Create a new http.Server struct.
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	// Use the server's ListenAndServe method to start the server
	log.Fatal(server.ListenAndServe())

}
