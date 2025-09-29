
# 🚄 Hamburg Rails

A Go-based microservice for representing a **directed weighted graph** (railway network), querying routes, and computing shortest paths.

## 📦 Getting Started

### Run the server

---
```bash
go run main.go --graph ./data/graph.txt
```
---

- The `--graph` flag (optional) loads an initial graph from a file.
- The server listens on **`:8080`** by default.

## 🧪 Testing

Run unit and fuzz tests:

---
```bash
# all tests
go test ./... -v

# fuzz testing example
go test ./internal/graphs -fuzz=Fuzz
```
---

## 🌐 REST API

### 1. Health check
---
```bash
curl -s http://localhost:8080/healthz
```
---
Response:
---
```json
{"status":"ok"}
```
---

### 2. Load graph (hot reload)
---
```bash
curl -X POST http://localhost:8080/admin/graph   -H "Content-Type: text/plain"   -d "AB5, BC4, CD8"
```
---
- Input format: comma-separated edges like `AB5` (edge from A→B with distance 5).
- Replaces the current graph in memory.

### 3. Get current graph
---
```bash
curl -s http://localhost:8080/graph | jq
```
---
Response:
---
```json
{
  "edges": {
    "A": [{"to":"B","distance":5}],
    "B": [{"to":"C","distance":4}],
    "C": [{"to":"D","distance":8}]
  },
  "node_count": 3
}
```
---

### 4. Distance for a fixed path
---
```bash
curl -X POST http://localhost:8080/routes/distance   -H "Content-Type: application/json"   -d '{"path":["A","B","C"]}'
```
---
Response:
---
```json
{"distance":9}
```
---

### 5. Count trips by stops
---
```bash
curl -X POST http://localhost:8080/routes/count-by-stops   -H "Content-Type: application/json"   -d '{"from":"A","to":"C","minStops":1,"maxStops":3}'
```
---
Response:
---
```json
{"count":1}
```
---

### 6. Count trips by distance
---
```bash
curl -X POST http://localhost:8080/routes/count-by-distance   -H "Content-Type: application/json"   -d '{"from":"A","to":"C","maxDistance":20}'
```
---
Response:
---
```json
{"count":1}
```
---

### 7. Shortest path
---
```bash
curl -s "http://localhost:8080/routes/shortest?from=A&to=C"
```
---
Response:
---
```json
{"distance":9,"path":["A","B","C"]}
```
---

### 8. Route Finder
---
```bash
curl -X POST http://localhost:8080/routes/search   -H "Content-Type: application/json"   -d '{"from": "A","to": "C","constraints": {"maxStops": 5,"maxDistance": 25,"distinctNodes": false},"limit": 10}'
```
---
Response:
---
```json
{"routes":[{"path":["A","B","C"],"distance":9}]}
```
---

## 📑 Architecture Decision Record (ADR)

### Context
We need to model a railway network as a directed weighted graph, support queries like:
- Distance for a given path
- Counting trips under constraints
- Finding shortest path

### Decision
- **Data structure**: adjacency list
---
```go
  map[string][]Edge
```
---
  where `Edge{To string; Distance int}`.
- **Shortest-path algorithm**: Dijkstra’s algorithm (with a heap).

### Consequences
- Efficient for sparse graphs.
- Supports fast route lookup.
- Time complexity:
  - Dijkstra: `O((V+E) log V)`
  - Distance query: `O(L)` for path length `L`
  - Trip counting (DFS): exponential in stops, bounded by constraints.
- Space complexity: `O(V+E)`.

