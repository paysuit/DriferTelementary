package main

import (
	"DriferTelementary/internal/routes"
	"log"
	"net/http"
)

// --- Main ---

func main() {

	router := routes.NewRouter()

	addr := ":8080"
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
