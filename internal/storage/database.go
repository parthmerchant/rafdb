package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Document represents a document in the database
type Document struct {
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Collection represents a collection of documents
type Collection struct {
	Name      string               `json:"name"`
	Documents map[string]*Document `json:"documents"`
	mu        sync.RWMutex
}

// Database represents the main database
type Database struct {
	Collections map[string]*Collection `json:"collections"`
	mu          sync.RWMutex
	dataFile    string
}

// NewDatabase creates a new database instance
func NewDatabase() *Database {
	return &Database{
		Collections: make(map[string]*Collection),
		dataFile:    "rafdb_data.json",
	}
}

// CreateCollection creates a new collection
func (db *Database) CreateCollection(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.Collections[name]; exists {
		return fmt.Errorf("collection '%s' already exists", name)
	}

	db.Collections[name] = &Collection{
		Name:      name,
		Documents: make(map[string]*Document),
	}

	return nil
}

// GetCollection returns a collection by name
func (db *Database) GetCollection(name string) (*Collection, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	collection, exists := db.Collections[name]
	if !exists {
		return nil, fmt.Errorf("collection '%s' not found", name)
	}

	return collection, nil
}

// ListCollections returns all collection names
func (db *Database) ListCollections() []string {
	db.mu.RLock()
	defer db.mu.RUnlock()

	names := make([]string, 0, len(db.Collections))
	for name := range db.Collections {
		names = append(names, name)
	}

	return names
}

// DeleteCollection deletes a collection
func (db *Database) DeleteCollection(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.Collections[name]; !exists {
		return fmt.Errorf("collection '%s' not found", name)
	}

	delete(db.Collections, name)
	return nil
}

// Insert inserts a document into a collection
func (c *Collection) Insert(id string, data map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.Documents[id]; exists {
		return fmt.Errorf("document with id '%s' already exists", id)
	}

	now := time.Now()
	c.Documents[id] = &Document{
		ID:        id,
		Data:      data,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return nil
}

// Get retrieves a document by ID
func (c *Collection) Get(id string) (*Document, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	doc, exists := c.Documents[id]
	if !exists {
		return nil, fmt.Errorf("document with id '%s' not found", id)
	}

	return doc, nil
}

// Update updates a document
func (c *Collection) Update(id string, data map[string]interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	doc, exists := c.Documents[id]
	if !exists {
		return fmt.Errorf("document with id '%s' not found", id)
	}

	doc.Data = data
	doc.UpdatedAt = time.Now()

	return nil
}

// Delete deletes a document
func (c *Collection) Delete(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.Documents[id]; !exists {
		return fmt.Errorf("document with id '%s' not found", id)
	}

	delete(c.Documents, id)
	return nil
}

// List returns all documents in the collection
func (c *Collection) List() []*Document {
	c.mu.RLock()
	defer c.mu.RUnlock()

	docs := make([]*Document, 0, len(c.Documents))
	for _, doc := range c.Documents {
		docs = append(docs, doc)
	}

	return docs
}

// Query performs a simple query on the collection
func (c *Collection) Query(field string, value interface{}) []*Document {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var results []*Document
	for _, doc := range c.Documents {
		if docValue, exists := doc.Data[field]; exists && docValue == value {
			results = append(results, doc)
		}
	}

	return results
}

// SaveToDisk saves the database to disk
func (db *Database) SaveToDisk() error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal database: %w", err)
	}

	err = os.WriteFile(db.dataFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write data file: %w", err)
	}

	return nil
}

// LoadFromDisk loads the database from disk
func (db *Database) LoadFromDisk() error {
	data, err := os.ReadFile(db.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, start with empty database
		}
		return fmt.Errorf("failed to read data file: %w", err)
	}

	var loadedDB Database
	err = json.Unmarshal(data, &loadedDB)
	if err != nil {
		return fmt.Errorf("failed to unmarshal database: %w", err)
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	db.Collections = loadedDB.Collections

	// Initialize mutexes for collections (they don't serialize)
	for _, collection := range db.Collections {
		collection.mu = sync.RWMutex{}
	}

	return nil
}

// Stats returns database statistics
func (db *Database) Stats() map[string]interface{} {
	db.mu.RLock()
	defer db.mu.RUnlock()

	stats := map[string]interface{}{
		"collections":     len(db.Collections),
		"total_documents": 0,
	}

	collectionStats := make(map[string]int)
	totalDocs := 0

	for name, collection := range db.Collections {
		collection.mu.RLock()
		docCount := len(collection.Documents)
		collection.mu.RUnlock()

		collectionStats[name] = docCount
		totalDocs += docCount
	}

	stats["total_documents"] = totalDocs
	stats["collection_stats"] = collectionStats

	return stats
}
