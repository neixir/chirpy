package main

import (
	"net/http"
)

// CH1 L4 https://www.boot.dev/lessons/861ada77-c583-42c8-a265-657f2c453103
// CH1 L5 https://www.boot.dev/lessons/8cf7315a-ffc0-4ce0-b482-5972ab695131

func main() {
	// 1. Create a new http.ServeMux
	// https://pkg.go.dev/net/http#NewServeMux
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))

	// CH1 L5
	// http: attempting to traverse a non-directory

	// 2. Create a new http.Server struct.
	server := http.Server{
		// 2.1 Use the new "ServeMux" as the server's handler
		Handler: mux,
		// 2.2 Set the .Addr field to ":8080"
		Addr: ":8080",
	}

	// 3. Use the server's ListenAndServe method to start the server
	server.ListenAndServe()

	// 4. Build and run your server (e.g. go build -o out && ./out)

	// 5. Open http://localhost:8080 in your browser.
	// You should see a 404 error because we haven't connected any handler logic yet.
	// Don't worry, that's what is expected for the tests to pass for now.
}
