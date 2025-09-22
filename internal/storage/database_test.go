package storage

import (
	"os"
	"testing"
)

func TestDatabase_CreateCollection(t *testing.T) {
	db := NewDatabase()

	err := db.CreateCollection("test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test duplicate collection
	err = db.CreateCollection("test")
	if err == nil {
		t.Fatal("Expected error for duplicate collection")
	}
}

func TestDatabase_GetCollection(t *testing.T) {
	db := NewDatabase()
	db.CreateCollection("test")

	collection, err := db.GetCollection("test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if collection.Name != "test" {
		t.Fatalf("Expected collection name 'test', got %s", collection.Name)
	}

	// Test non-existent collection
	_, err = db.GetCollection("nonexistent")
	if err == nil {
		t.Fatal("Expected error for non-existent collection")
	}
}

func TestCollection_Insert(t *testing.T) {
	db := NewDatabase()
	db.CreateCollection("test")
	collection, _ := db.GetCollection("test")

	data := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	err := collection.Insert("user1", data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test duplicate ID
	err = collection.Insert("user1", data)
	if err == nil {
		t.Fatal("Expected error for duplicate ID")
	}
}

func TestCollection_Get(t *testing.T) {
	db := NewDatabase()
	db.CreateCollection("test")
	collection, _ := db.GetCollection("test")

	data := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	collection.Insert("user1", data)

	doc, err := collection.Get("user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if doc.ID != "user1" {
		t.Fatalf("Expected ID 'user1', got %s", doc.ID)
	}

	if doc.Data["name"] != "John" {
		t.Fatalf("Expected name 'John', got %v", doc.Data["name"])
	}

	// Test non-existent document
	_, err = collection.Get("nonexistent")
	if err == nil {
		t.Fatal("Expected error for non-existent document")
	}
}

func TestCollection_Update(t *testing.T) {
	db := NewDatabase()
	db.CreateCollection("test")
	collection, _ := db.GetCollection("test")

	originalData := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	collection.Insert("user1", originalData)

	updatedData := map[string]interface{}{
		"name": "John Doe",
		"age":  31,
	}

	err := collection.Update("user1", updatedData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	doc, _ := collection.Get("user1")
	if doc.Data["name"] != "John Doe" {
		t.Fatalf("Expected updated name 'John Doe', got %v", doc.Data["name"])
	}

	// Test non-existent document
	err = collection.Update("nonexistent", updatedData)
	if err == nil {
		t.Fatal("Expected error for non-existent document")
	}
}

func TestCollection_Delete(t *testing.T) {
	db := NewDatabase()
	db.CreateCollection("test")
	collection, _ := db.GetCollection("test")

	data := map[string]interface{}{
		"name": "John",
		"age":  30,
	}

	collection.Insert("user1", data)

	err := collection.Delete("user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify document is deleted
	_, err = collection.Get("user1")
	if err == nil {
		t.Fatal("Expected error for deleted document")
	}

	// Test non-existent document
	err = collection.Delete("nonexistent")
	if err == nil {
		t.Fatal("Expected error for non-existent document")
	}
}

func TestCollection_Query(t *testing.T) {
	db := NewDatabase()
	db.CreateCollection("test")
	collection, _ := db.GetCollection("test")

	// Insert test data
	collection.Insert("user1", map[string]interface{}{
		"name": "John",
		"age":  30,
		"city": "New York",
	})

	collection.Insert("user2", map[string]interface{}{
		"name": "Jane",
		"age":  25,
		"city": "New York",
	})

	collection.Insert("user3", map[string]interface{}{
		"name": "Bob",
		"age":  35,
		"city": "San Francisco",
	})

	// Query by city
	results := collection.Query("city", "New York")
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Query by age
	results = collection.Query("age", 30)
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].Data["name"] != "John" {
		t.Fatalf("Expected name 'John', got %v", results[0].Data["name"])
	}
}

func TestDatabase_Persistence(t *testing.T) {
	// Use a temporary file for testing
	tempFile := "test_rafdb_data.json"
	defer os.Remove(tempFile)

	// Create database and add data
	db := NewDatabase()
	db.dataFile = tempFile

	db.CreateCollection("users")
	collection, _ := db.GetCollection("users")
	collection.Insert("user1", map[string]interface{}{
		"name": "John",
		"age":  30,
	})

	// Save to disk
	err := db.SaveToDisk()
	if err != nil {
		t.Fatalf("Expected no error saving to disk, got %v", err)
	}

	// Create new database and load from disk
	db2 := NewDatabase()
	db2.dataFile = tempFile

	err = db2.LoadFromDisk()
	if err != nil {
		t.Fatalf("Expected no error loading from disk, got %v", err)
	}

	// Verify data was loaded
	collection2, err := db2.GetCollection("users")
	if err != nil {
		t.Fatalf("Expected collection to exist after loading, got %v", err)
	}

	doc, err := collection2.Get("user1")
	if err != nil {
		t.Fatalf("Expected document to exist after loading, got %v", err)
	}

	if doc.Data["name"] != "John" {
		t.Fatalf("Expected name 'John' after loading, got %v", doc.Data["name"])
	}
}

func TestDatabase_Stats(t *testing.T) {
	db := NewDatabase()

	// Create collections and add documents
	db.CreateCollection("users")
	db.CreateCollection("products")

	users, _ := db.GetCollection("users")
	products, _ := db.GetCollection("products")

	users.Insert("user1", map[string]interface{}{"name": "John"})
	users.Insert("user2", map[string]interface{}{"name": "Jane"})

	products.Insert("prod1", map[string]interface{}{"name": "Laptop"})

	stats := db.Stats()

	if stats["collections"] != 2 {
		t.Fatalf("Expected 2 collections, got %v", stats["collections"])
	}

	if stats["total_documents"] != 3 {
		t.Fatalf("Expected 3 total documents, got %v", stats["total_documents"])
	}

	collectionStats := stats["collection_stats"].(map[string]int)
	if collectionStats["users"] != 2 {
		t.Fatalf("Expected 2 users, got %d", collectionStats["users"])
	}

	if collectionStats["products"] != 1 {
		t.Fatalf("Expected 1 product, got %d", collectionStats["products"])
	}
}

func TestConcurrentAccess(t *testing.T) {
	db := NewDatabase()
	db.CreateCollection("concurrent")
	collection, _ := db.GetCollection("concurrent")

	// Test concurrent writes
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			data := map[string]interface{}{
				"id":    id,
				"value": id * 10,
			}
			collection.Insert(string(rune('a'+id)), data)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all documents were inserted
	docs := collection.List()
	if len(docs) != 10 {
		t.Fatalf("Expected 10 documents, got %d", len(docs))
	}
}

func BenchmarkInsert(b *testing.B) {
	db := NewDatabase()
	db.CreateCollection("benchmark")
	collection, _ := db.GetCollection("benchmark")

	data := map[string]interface{}{
		"name":  "Test User",
		"email": "test@example.com",
		"age":   25,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collection.Insert(string(rune(i)), data)
	}
}

func BenchmarkGet(b *testing.B) {
	db := NewDatabase()
	db.CreateCollection("benchmark")
	collection, _ := db.GetCollection("benchmark")

	// Pre-populate with data
	data := map[string]interface{}{
		"name":  "Test User",
		"email": "test@example.com",
		"age":   25,
	}

	for i := 0; i < 1000; i++ {
		collection.Insert(string(rune(i)), data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collection.Get(string(rune(i % 1000)))
	}
}

func BenchmarkQuery(b *testing.B) {
	db := NewDatabase()
	db.CreateCollection("benchmark")
	collection, _ := db.GetCollection("benchmark")

	// Pre-populate with data
	cities := []string{"New York", "San Francisco", "Chicago", "Boston", "Seattle"}

	for i := 0; i < 1000; i++ {
		data := map[string]interface{}{
			"name": "User " + string(rune(i)),
			"age":  20 + (i % 50),
			"city": cities[i%len(cities)],
		}
		collection.Insert(string(rune(i)), data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collection.Query("city", cities[i%len(cities)])
	}
}
