// cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/oxtx/go-crdt-sync/internal/api"
	"github.com/oxtx/go-crdt-sync/internal/service"
	"github.com/oxtx/go-crdt-sync/internal/store"
)

func main() {
	mem := store.NewMemoryStore()
	svc := service.New(mem)

	r := chi.NewRouter()
	api.Attach(r, svc)

	addr := getEnv("ADDR", ":8080")
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
