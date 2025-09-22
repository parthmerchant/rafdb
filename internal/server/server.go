package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"rafdb/internal/storage"
)

// Server represents the HTTP server
type Server struct {
	db     *storage.Database
	server *http.Server
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewServer creates a new server instance
func NewServer(db *storage.Database) *Server {
	return &Server{
		db: db,
	}
}

// Start starts the HTTP server
func (s *Server) Start(addr string) {
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Collection routes
	api.HandleFunc("/collections", s.handleListCollections).Methods("GET")
	api.HandleFunc("/collections", s.handleCreateCollection).Methods("POST")
	api.HandleFunc("/collections/{collection}", s.handleDeleteCollection).Methods("DELETE")

	// Document routes
	api.HandleFunc("/collections/{collection}/documents", s.handleListDocuments).Methods("GET")
	api.HandleFunc("/collections/{collection}/documents", s.handleInsertDocument).Methods("POST")
	api.HandleFunc("/collections/{collection}/documents/{id}", s.handleGetDocument).Methods("GET")
	api.HandleFunc("/collections/{collection}/documents/{id}", s.handleUpdateDocument).Methods("PUT")
	api.HandleFunc("/collections/{collection}/documents/{id}", s.handleDeleteDocument).Methods("DELETE")

	// Query route
	api.HandleFunc("/collections/{collection}/query", s.handleQuery).Methods("POST")

	// Stats route
	api.HandleFunc("/stats", s.handleStats).Methods("GET")

	// Health check
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	handler := c.Handler(router)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Fatal(s.server.ListenAndServe())
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}

// Helper function to send JSON response
func (s *Server) sendResponse(w http.ResponseWriter, success bool, data interface{}, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")

	response := Response{
		Success: success,
		Data:    data,
		Error:   errorMsg,
	}

	if !success {
		w.WriteHeader(http.StatusBadRequest)
	}

	json.NewEncoder(w).Encode(response)
}

// Collection handlers
func (s *Server) handleListCollections(w http.ResponseWriter, r *http.Request) {
	collections := s.db.ListCollections()
	s.sendResponse(w, true, collections, "")
}

func (s *Server) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendResponse(w, false, nil, "Invalid JSON")
		return
	}

	if req.Name == "" {
		s.sendResponse(w, false, nil, "Collection name is required")
		return
	}

	if err := s.db.CreateCollection(req.Name); err != nil {
		s.sendResponse(w, false, nil, err.Error())
		return
	}

	s.sendResponse(w, true, map[string]string{"message": "Collection created successfully"}, "")
}

func (s *Server) handleDeleteCollection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectionName := vars["collection"]

	if err := s.db.DeleteCollection(collectionName); err != nil {
		s.sendResponse(w, false, nil, err.Error())
		return
	}

	s.sendResponse(w, true, map[string]string{"message": "Collection deleted successfully"}, "")
}

// Document handlers
func (s *Server) handleListDocuments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectionName := vars["collection"]

	collection, err := s.db.GetCollection(collectionName)
	if err != nil {
		s.sendResponse(w, false, nil, err.Error())
		return
	}

	documents := collection.List()
	s.sendResponse(w, true, documents, "")
}

func (s *Server) handleInsertDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectionName := vars["collection"]

	collection, err := s.db.GetCollection(collectionName)
	if err != nil {
		// Try to create the collection if it doesn't exist
		if err := s.db.CreateCollection(collectionName); err != nil {
			s.sendResponse(w, false, nil, err.Error())
			return
		}
		collection, _ = s.db.GetCollection(collectionName)
	}

	var req struct {
		ID   string                 `json:"id"`
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendResponse(w, false, nil, "Invalid JSON")
		return
	}

	if req.ID == "" {
		s.sendResponse(w, false, nil, "Document ID is required")
		return
	}

	if err := collection.Insert(req.ID, req.Data); err != nil {
		s.sendResponse(w, false, nil, err.Error())
		return
	}

	s.sendResponse(w, true, map[string]string{"message": "Document inserted successfully"}, "")
}

func (s *Server) handleGetDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectionName := vars["collection"]
	documentID := vars["id"]

	collection, err := s.db.GetCollection(collectionName)
	if err != nil {
		s.sendResponse(w, false, nil, err.Error())
		return
	}

	document, err := collection.Get(documentID)
	if err != nil {
		s.sendResponse(w, false, nil, err.Error())
		return
	}

	s.sendResponse(w, true, document, "")
}

func (s *Server) handleUpdateDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectionName := vars["collection"]
	documentID := vars["id"]

	collection, err := s.db.GetCollection(collectionName)
	if err != nil {
		s.sendResponse(w, false, nil, err.Error())
		return
	}

	var req struct {
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendResponse(w, false, nil, "Invalid JSON")
		return
	}

	if err := collection.Update(documentID, req.Data); err != nil {
		s.sendResponse(w, false, nil, err.Error())
		return
	}

	s.sendResponse(w, true, map[string]string{"message": "Document updated successfully"}, "")
}

func (s *Server) handleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectionName := vars["collection"]
	documentID := vars["id"]

	collection, err := s.db.GetCollection(collectionName)
	if err != nil {
		s.sendResponse(w, false, nil, err.Error())
		return
	}

	if err := collection.Delete(documentID); err != nil {
		s.sendResponse(w, false, nil, err.Error())
		return
	}

	s.sendResponse(w, true, map[string]string{"message": "Document deleted successfully"}, "")
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectionName := vars["collection"]

	collection, err := s.db.GetCollection(collectionName)
	if err != nil {
		s.sendResponse(w, false, nil, err.Error())
		return
	}

	var req struct {
		Field string      `json:"field"`
		Value interface{} `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendResponse(w, false, nil, "Invalid JSON")
		return
	}

	if req.Field == "" {
		s.sendResponse(w, false, nil, "Field is required for query")
		return
	}

	results := collection.Query(req.Field, req.Value)
	s.sendResponse(w, true, results, "")
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.db.Stats()
	s.sendResponse(w, true, stats, "")
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.sendResponse(w, true, map[string]string{
		"status":  "healthy",
		"version": "1.0.0",
		"name":    "RAFDB",
	}, "")
}
