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
	mux := http.NewServeMux()
	handlers := Handlers{store: kvStore}

	mux.HandleFunc("PUT /kv/{key}", handlers.Put)

	return &http.Server{
		Addr:     address,
		Handler:  mux,
		ErrorLog: errorLog,
	}
}
