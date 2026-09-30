package httpapi

import (
	"io"
	"log"
	"net/http"

	"github.com/Aditya090202/ark-kv/internal/store"
)

// Handlers will contain the HTTP endpoint dependencies.
type Handlers struct {
	// these are dependencies for the handlers
	// hence why they are called "dependency injections"
	// store is here to manipulate the state of the store in a concurrent safe way
	// the logger is to record the deletion of a record, as the delete is idempotent
	// meaning it will return a 200 OK response regardless of if the key existed or not
	// all it is concerned about is the "final state" of the state of the store
	// the different states are defined as either having the key or not having the key in the store
	store  *store.Store
	logger *log.Logger
}

// TODO: Add PUT, GET, DELETE, clear, and health handlers.
// TODO: Return the walkthrough's required 400, 404, and 405 responses.

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) Put(w http.ResponseWriter, r *http.Request) {
	log.Printf(
		"accepted method=%s path=%s remote=%s userAgent=%q",
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		r.UserAgent(),
	)
	// get the wildcard value which is the key, using the PathValue function
	key := r.PathValue("key")
	// if key is not passed in to the request, return a 400
	if len(key) == 0 {
		http.Error(w, "key cannot be empty", http.StatusBadRequest)
		return
	}
	// body is a readable stream, you can read through it using a io.ReadAll
	body, err := io.ReadAll(r.Body)
	// Error reading the body
	if err != nil {
		http.Error(w, "could not read from the stream", http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/plain")
		return
	}
	// body is empty
	if len(body) == 0 {
		http.Error(w, "value cannot be empty", http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/plain")
		return
	}
	//store value and key in the store
	value := string(body)
	h.store.Set(key, value)
	// set the header "Content-Type to text/plain; charset=utf-8"
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

}
func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	log.Printf(
		"accepted method=%s path=%s remote=%s userAgent=%q",
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		r.UserAgent(),
	)
	key := r.PathValue("key")
	// if key is not passed in to the request, return a 400
	if len(key) == 0 {
		http.Error(w, "key cannot be empty", http.StatusBadRequest)
		return
	}
	value, exists := h.store.Get(key)
	if !exists {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, value)

}

// idempotent delete from the store
func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if len(key) == 0 {
		http.Error(w, "key cannot be empty", http.StatusBadRequest)
		return
	}
	existed := h.store.Delete(key)
	h.logger.Printf("DELETE key=%q existed=%t", key, existed)
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) Clear_All(w http.ResponseWriter, r *http.Request) {
	h.store.Clear_All()
	w.WriteHeader(http.StatusOK)
}
