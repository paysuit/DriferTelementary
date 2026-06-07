package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// --- Models ---

type Item struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateItemRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type APIResponse struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// --- In-memory store ---

type Store struct {
	mu     sync.RWMutex
	items  map[int]Item
	nextID int
}

func NewStore() *Store {
	s := &Store{
		items:  make(map[int]Item),
		nextID: 1,
	}
	// Seed mock data
	mockItems := []CreateItemRequest{
		{Name: "Wireless Mouse", Price: 29.99},
		{Name: "Mechanical Keyboard", Price: 89.99},
		{Name: "USB-C Hub", Price: 49.99},
		{Name: "Monitor Stand", Price: 34.99},
		{Name: "Webcam HD", Price: 74.99},
	}
	for _, m := range mockItems {
		s.Create(m.Name, m.Price)
	}
	return s
}

func (s *Store) GetAll() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	return items
}

func (s *Store) GetByID(id int) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	return item, ok
}

func (s *Store) Create(name string, price float64) Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	item := Item{
		ID:        s.nextID,
		Name:      name,
		Price:     price,
		CreatedAt: time.Now(),
	}
	s.items[s.nextID] = item
	s.nextID++
	return item
}

func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}

// --- Handlers ---

type Server struct {
	store *Store
}

func NewServer(store *Store) *Server {
	return &Server{store: store}
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// GET /items — list all items
func (s *Server) handleListItems(w http.ResponseWriter, r *http.Request) {
	items := s.store.GetAll()
	writeJSON(w, http.StatusOK, APIResponse{Data: items})
}

// POST /items — create a new item
func (s *Server) handleCreateItem(w http.ResponseWriter, r *http.Request) {
	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Error: "invalid request body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{Error: "name is required"})
		return
	}
	if req.Price < 0 {
		writeJSON(w, http.StatusBadRequest, APIResponse{Error: "price must be non-negative"})
		return
	}
	item := s.store.Create(req.Name, req.Price)
	writeJSON(w, http.StatusCreated, APIResponse{Data: item, Message: "item created"})
}

// GET /items/{id} — get a single item
func (s *Server) handleGetItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Error: "invalid id"})
		return
	}
	item, ok := s.store.GetByID(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, APIResponse{Error: "item not found"})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Data: item})
}

// DELETE /items/{id} — delete an item
func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{Error: "invalid id"})
		return
	}
	if !s.store.Delete(id) {
		writeJSON(w, http.StatusNotFound, APIResponse{Error: "item not found"})
		return
	}
	writeJSON(w, http.StatusOK, APIResponse{Message: "item deleted"})
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

// --- Router ---

// itemsRouter routes /items and /items/{id} based on method and path.
func (s *Server) itemsRouter(w http.ResponseWriter, r *http.Request) {
	// /items (no trailing ID)
	if r.URL.Path == "/items" || r.URL.Path == "/items/" {
		switch r.Method {
		case http.MethodGet:
			s.handleListItems(w, r)
		case http.MethodPost:
			s.handleCreateItem(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
		}
		return
	}

	// /items/{id}
	switch r.Method {
	case http.MethodGet:
		s.handleGetItem(w, r)
	case http.MethodDelete:
		s.handleDeleteItem(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{Error: "method not allowed"})
	}
}

// --- Middleware ---

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s — %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// --- Helpers ---

// parseID extracts the trailing integer from a path like /items/42.
func parseID(path string) (int, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return 0, fmt.Errorf("no id in path")
	}
	return strconv.Atoi(parts[len(parts)-1])
}

// --- Main ---

func main() {
	store := NewStore()
	server := NewServer(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/items", server.itemsRouter)
	mux.HandleFunc("/items/", server.itemsRouter)

	handler := loggingMiddleware(mux)

	addr := ":8080"
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
