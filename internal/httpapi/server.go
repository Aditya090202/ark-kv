// Package httpapi will expose the key-value store over HTTP.
package httpapi

import (
	"log"
	"net/http"

	"github.com/Aditya090202/ark-kv/internal/store"
)

// // Server will own HTTP server configuration and route registration.
// type Server struct {
// 	httpServer *http.Server
// 	address    string
// 	Handler    http.Handler
// 	ErrorLog   *log.Logger
// }

// TODO: Add server construction, route registration, and startup behavior.
// create a method that defines a server
func NewServer(address string, kvStore *store.Store, errorLog *log.Logger) *http.Server {
	//create a request multiplexer that handles matching a pattern to a specific handler
	mux := http.NewServeMux()
	// this is an instance of the Handlers type inside handlers.go
	// it will contain all the handlers that we need for this project
	//define the store dependency that is already passed in
	handlers := Handlers{store: kvStore}

	// creating the handlers for this server
	mux.HandleFunc("PUT /kv/{key}", handlers.Put)
	mux.HandleFunc("GET /kv/{key}", handlers.Get)
	mux.HandleFunc("GET /health", handlers.Health)

	// return said http.Server pointer with respective inputs
	// NOTE: You can use the mux as a value for Handler interface, as the ServeMux type matches the interface
	return &http.Server{
		Addr:     address,
		Handler:  mux,
		ErrorLog: errorLog,
	}
}
