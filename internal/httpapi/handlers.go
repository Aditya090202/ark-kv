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

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) Put(w http.ResponseWriter, r *http.Request) {
	// get the wildcard value which is the key, using the PathValue function
	key := r.PathValue("key")
	// body is a readable stream, you can read through it using a io.ReadAll
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read from the stream", http.StatusBadRequest)
		return
	}
	value := string(body)
	h.store.Set(key, value)

	log.Printf(
		"accepted method=%s path=%s remote=%s userAgent=%q",
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		r.UserAgent(),
	)
}
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	value, exists := h.store.Get(key)
	if !exists {
		http.Error(w, "key not found", http.StatusBadRequest)
	}
	io.WriteString(w, value+"\n")

	log.Printf(
		"accepted method=%s path=%s remote=%s userAgent=%q",
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		r.UserAgent(),
	)
}
