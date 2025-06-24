package main

import (
	"log"
	"net/http"
)

// CH1 L04 https://www.boot.dev/lessons/861ada77-c583-42c8-a265-657f2c453103
// CH1 L05 https://www.boot.dev/lessons/8cf7315a-ffc0-4ce0-b482-5972ab695131
// CH1 L08 https://www.boot.dev/lessons/20709716-4d7c-47fe-b182-9bccf8436ddc

func main() {
	// CH1 L4-L5
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.Handle("/assets", http.FileServer(http.Dir("./assets"))) // CH1 L05

	// Create a new http.Server struct.
	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	// Use the server's ListenAndServe method to start the server
	log.Fatal(server.ListenAndServe())

}
