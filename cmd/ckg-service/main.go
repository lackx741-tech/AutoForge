package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type CodeEntity struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"` // function, class, module
	Name         string    `json:"name"`
	FilePath     string    `json:"file_path"`
	Content      string    `json:"content"`
	Embedding    []float64 `json:"embedding,omitempty"`
	Dependencies []string  `json:"dependencies"`
}

var db *sql.DB

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./ckg.db")
	if err != nil {
		log.Fatal(err)
	}

	schema := `CREATE TABLE IF NOT EXISTS entities (
		id TEXT PRIMARY KEY,
		type TEXT,
		name TEXT,
		file_path TEXT,
		content TEXT,
		dependencies TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_name ON entities(name);
	CREATE INDEX IF NOT EXISTS idx_type ON entities(type);`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}
}

func ingestHandler(w http.ResponseWriter, r *http.Request) {
	var entity CodeEntity
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	deps, _ := json.Marshal(entity.Dependencies)
	_, err := db.Exec(
		"INSERT OR REPLACE INTO entities VALUES (?, ?, ?, ?, ?, ?, datetime('now'))",
		entity.ID, entity.Type, entity.Name, entity.FilePath, entity.Content, string(deps),
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ingested", "id": entity.ID})
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	rows, err := db.Query(
		"SELECT id, type, name, file_path FROM entities WHERE name LIKE ?",
		"%"+name+"%",
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	var results []CodeEntity
	for rows.Next() {
		var e CodeEntity
		rows.Scan(&e.ID, &e.Type, &e.Name, &e.FilePath)
		results = append(results, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func main() {
	initDB()
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ingest", ingestHandler)
	http.HandleFunc("/query", queryHandler)

	log.Printf("CKG Service on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
