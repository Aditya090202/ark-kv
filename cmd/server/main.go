// Command server will run a single key-value store node.
package main

import (
	"fmt"
	"log"
	"net/http"
)

// TODO: Construct the store and HTTP server, then listen on port 8080.

func main() {
	mux := http.NewServeMux()
	keyHandler := func(resWriter http.ResponseWriter, r *http.Request) {
		fmt.Println("Hello, world!")
	}
	logRequestArrivalHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf(
			"accepted method=%s path=%s remote=%s userAgent=%q",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			r.UserAgent(),
		)
		w.WriteHeader(http.StatusOK)
	}
	mux.HandleFunc("GET /key", keyHandler)
	mux.HandleFunc("GET /", logRequestArrivalHandler)

	fmt.Println("Server is running.....")
	log.Fatal(http.ListenAndServe(":8080", mux))

}
