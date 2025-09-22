#!/bin/bash

# RAFDB Basic Usage Example
# This script demonstrates basic RAFDB operations

echo "🐕 RAFDB Basic Usage Example"
echo "================================"

# Check if RAFDB is running
if ! curl -s http://localhost:8080/api/v1/health > /dev/null; then
    echo "❌ RAFDB is not running. Please start it first with 'just run' or 'just deploy-rafdb'"
    exit 1
fi

echo "✅ RAFDB is running!"
echo ""

# 1. Health Check
echo "1. 🏥 Health Check:"
curl -s http://localhost:8080/api/v1/health | jq .
echo ""

# 2. Create a collection
echo "2. 📁 Creating 'books' collection:"
curl -s -X POST http://localhost:8080/api/v1/collections \
    -H "Content-Type: application/json" \
    -d '{"name": "books"}' | jq .
echo ""

# 3. Insert some books
echo "3. 📚 Inserting books:"

curl -s -X POST http://localhost:8080/api/v1/collections/books/documents \
    -H "Content-Type: application/json" \
    -d '{
        "id": "book1", 
        "data": {
            "title": "The Go Programming Language",
            "author": "Alan Donovan",
            "year": 2015,
            "genre": "Programming",
            "available": true
        }
    }' | jq .

curl -s -X POST http://localhost:8080/api/v1/collections/books/documents \
    -H "Content-Type: application/json" \
    -d '{
        "id": "book2", 
        "data": {
            "title": "Clean Code",
            "author": "Robert Martin",
            "year": 2008,
            "genre": "Programming",
            "available": false
        }
    }' | jq .

curl -s -X POST http://localhost:8080/api/v1/collections/books/documents \
    -H "Content-Type: application/json" \
    -d '{
        "id": "book3", 
        "data": {
            "title": "1984",
            "author": "George Orwell",
            "year": 1949,
            "genre": "Fiction",
            "available": true
        }
    }' | jq .

echo ""

# 4. Get a specific book
echo "4. 📖 Getting book1:"
curl -s http://localhost:8080/api/v1/collections/books/documents/book1 | jq .
echo ""

# 5. List all books
echo "5. 📋 Listing all books:"
curl -s http://localhost:8080/api/v1/collections/books/documents | jq .
echo ""

# 6. Query books by genre
echo "6. 🔍 Finding all Programming books:"
curl -s -X POST http://localhost:8080/api/v1/collections/books/query \
    -H "Content-Type: application/json" \
    -d '{"field": "genre", "value": "Programming"}' | jq .
echo ""

# 7. Query available books
echo "7. ✅ Finding all available books:"
curl -s -X POST http://localhost:8080/api/v1/collections/books/query \
    -H "Content-Type: application/json" \
    -d '{"field": "available", "value": true}' | jq .
echo ""

# 8. Update a book
echo "8. ✏️  Updating book2 (making it available):"
curl -s -X PUT http://localhost:8080/api/v1/collections/books/documents/book2 \
    -H "Content-Type: application/json" \
    -d '{
        "data": {
            "title": "Clean Code",
            "author": "Robert Martin",
            "year": 2008,
            "genre": "Programming",
            "available": true,
            "condition": "excellent"
        }
    }' | jq .
echo ""

# 9. Database statistics
echo "9. 📊 Database statistics:"
curl -s http://localhost:8080/api/v1/stats | jq .
echo ""

# 10. List all collections
echo "10. 📂 All collections:"
curl -s http://localhost:8080/api/v1/collections | jq .
echo ""

echo "🎉 Example completed! Your data has been persisted to rafdb_data.json"
echo "You can restart RAFDB and your data will still be there!"
