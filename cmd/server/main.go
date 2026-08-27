package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Aditya090202/distributed-key-value-store/internal/server"
)

func main() {
	port := flag.Int("port", 6379, "TCP port to listen on")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}

	log.Printf("miniKV listening on %s (Redis-compatible RESP)", addr)

	srv := server.New()

	// Accept connections in a goroutine so we can cleanly shut down
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// listener closed via signal — expected during shutdown
				return
			}
			go srv.HandleConnection(conn)
		}
	}()

	// Wait for SIGINT/SIGTERM, then shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("shutting down...")
	listener.Close()
	srv.Shutdown()
}