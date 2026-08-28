// Command server will run a single key-value store node.
package main

import (
	"fmt"
	"log"

	"github.com/Aditya090202/ark-kv/internal/httpapi"
	"github.com/Aditya090202/ark-kv/internal/store"
)

// TODO: Construct the store and HTTP server, then listen on port 8080.

func main() {
	kvStore := store.NewStore()
	errorLog := log.Default()

	server := httpapi.NewServer(":8080", kvStore, errorLog)

	fmt.Println("Server has started running on port 8080......")
	errorLog.Fatal(server.ListenAndServe())
}
