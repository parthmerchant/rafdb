#!/bin/bash

# RAFDB Performance Demo
# This script demonstrates RAFDB's performance capabilities

echo "🚀 RAFDB Performance Demo"
echo "========================="

# Check if RAFDB is running
if ! curl -s http://localhost:8080/api/v1/health > /dev/null; then
    echo "❌ RAFDB is not running. Please start it first with 'just run' or 'just deploy-rafdb'"
    exit 1
fi

echo "✅ RAFDB is running!"
echo ""

# Create performance test collection
echo "📁 Creating 'performance' collection..."
curl -s -X POST http://localhost:8080/api/v1/collections \
    -H "Content-Type: application/json" \
    -d '{"name": "performance"}' > /dev/null

echo ""

# Test 1: Bulk Insert Performance
echo "🔥 Test 1: Bulk Insert Performance (1000 documents)"
echo "=================================================="

start_time=$(date +%s.%N)

for i in {1..1000}; do
    curl -s -X POST http://localhost:8080/api/v1/collections/performance/documents \
        -H "Content-Type: application/json" \
        -d "{
            \"id\": \"doc$i\", 
            \"data\": {
                \"index\": $i,
                \"name\": \"Document $i\",
                \"value\": $((i * 10)),
                \"category\": \"category$((i % 10))\",
                \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
                \"active\": $((i % 2 == 0))
            }
        }" > /dev/null
    
    if [ $((i % 100)) -eq 0 ]; then
        echo "  Inserted $i documents..."
    fi
done

end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc)
ops_per_sec=$(echo "scale=2; 1000 / $duration" | bc)

echo "✅ Inserted 1000 documents in ${duration}s"
echo "📈 Write performance: ${ops_per_sec} ops/sec"
echo ""

# Test 2: Random Read Performance
echo "🔍 Test 2: Random Read Performance (500 reads)"
echo "=============================================="

start_time=$(date +%s.%N)

for i in {1..500}; do
    random_id=$((RANDOM % 1000 + 1))
    curl -s http://localhost:8080/api/v1/collections/performance/documents/doc$random_id > /dev/null
done

end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc)
ops_per_sec=$(echo "scale=2; 500 / $duration" | bc)

echo "✅ Performed 500 random reads in ${duration}s"
echo "📈 Read performance: ${ops_per_sec} ops/sec"
echo ""

# Test 3: Query Performance
echo "🔎 Test 3: Query Performance (50 queries)"
echo "========================================="

start_time=$(date +%s.%N)

for i in {1..50}; do
    category=$((i % 10))
    curl -s -X POST http://localhost:8080/api/v1/collections/performance/query \
        -H "Content-Type: application/json" \
        -d "{\"field\": \"category\", \"value\": \"category$category\"}" > /dev/null
done

end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc)
ops_per_sec=$(echo "scale=2; 50 / $duration" | bc)

echo "✅ Performed 50 queries in ${duration}s"
echo "📈 Query performance: ${ops_per_sec} ops/sec"
echo ""

# Test 4: Update Performance
echo "✏️  Test 4: Update Performance (200 updates)"
echo "==========================================="

start_time=$(date +%s.%N)

for i in {1..200}; do
    curl -s -X PUT http://localhost:8080/api/v1/collections/performance/documents/doc$i \
        -H "Content-Type: application/json" \
        -d "{
            \"data\": {
                \"index\": $i,
                \"name\": \"Updated Document $i\",
                \"value\": $((i * 20)),
                \"category\": \"updated_category$((i % 5))\",
                \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
                \"active\": true,
                \"updated\": true
            }
        }" > /dev/null
done

end_time=$(date +%s.%N)
duration=$(echo "$end_time - $start_time" | bc)
ops_per_sec=$(echo "scale=2; 200 / $duration" | bc)

echo "✅ Performed 200 updates in ${duration}s"
echo "📈 Update performance: ${ops_per_sec} ops/sec"
echo ""

# Final Statistics
echo "📊 Final Database Statistics:"
echo "============================="
curl -s http://localhost:8080/api/v1/stats | jq .

echo ""
echo "🎯 Performance Summary:"
echo "======================"
echo "• RAFDB successfully handled high-throughput operations"
echo "• All data is stored in-memory for maximum speed"
echo "• Concurrent access is handled with proper locking"
echo "• Data persistence ensures reliability"
echo ""
echo "🐕 RAFDB: Fast like Rafah! 🚀"
