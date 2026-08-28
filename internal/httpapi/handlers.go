package httpapi

import (
	"io"
	"log"
	"net/http"

	"github.com/Aditya090202/ark-kv/internal/store"
)

// Handlers will contain the HTTP endpoint dependencies.
type Handlers struct {
	store *store.Store
}

// TODO: Add PUT, GET, DELETE, clear, and health handlers.
// TODO: Return the walkthrough's required 400, 404, and 405 responses.

func (h *Handlers) Put(w http.ResponseWriter, r *http.Request) {
	log.Printf(
		"accepted method=%s path=%s remote=%s userAgent=%q",
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		r.UserAgent(),
	)
	io.WriteString(w, "This should set a key to the store\n")
}
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	log.Printf(
		"accepted method=%s path=%s remote=%s userAgent=%q",
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		r.UserAgent(),
	)
	io.WriteString(w, "This should get a key from the store if it exists")
}
