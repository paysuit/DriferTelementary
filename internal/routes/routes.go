package routes

import (
	"DriferTelementary/internal/websocket"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type APIResponse struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// GET / — root, lists available routes
func handleRoot(w http.ResponseWriter, r *http.Request) {
	info := map[string]any{
		"name":    "Items API",
		"version": "1.0.0",
		"routes": []map[string]string{
			{"method": "GET", "path": "/", "description": "API info"},
			{"method": "GET", "path": "/health", "description": "Health check"},
			{"method": "GET", "path": "/items", "description": "List all items"},
			{"method": "POST", "path": "/items", "description": "Create an item"},
			{"method": "GET", "path": "/items/{id}", "description": "Get item by ID"},
			{"method": "DELETE", "path": "/items/{id}", "description": "Delete item by ID"},
		},
	}
	writeJSON(w, http.StatusOK, APIResponse{Data: info})
}

// GET /health — health check
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, APIResponse{Message: "ok"})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s — %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func NewRouter() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/health", handleHealth)

	hub := websocket.NewHub()
	wsHandler := websocket.NewHandler(hub)
	mux.Handle("/ws", wsHandler)

	return loggingMiddleware(mux)

}
