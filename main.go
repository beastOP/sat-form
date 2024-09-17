package main

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"sat-form/database"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	db, err := sql.Open("sqlite3", "./sat_scores.db")
	if err != nil {
		log.Fatalf("Error: %+v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Error: %+v", err)
	}
	log.Println("Connected to database")
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatalf("Error: %+v", err)
	}

	err = goose.Up(db, "migrations")
	if err != nil {
		log.Fatalf("Error: %+v", err)
	}

	_ = database.New(db)

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", handleIndexRoute)

	apiMux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	apiMux.HandleFunc("/submit-sat-score", handleSubmitSATForm)
	apiMux.HandleFunc("/get-rank", handleGetRank)
	apiMux.HandleFunc("/update-sat-score", handleUpdateSATScore)
	apiMux.HandleFunc("/delete-record", handleDeleteRecord)
	apiMux.HandleFunc("/view-all-data", handleViewAllData)

	fmt.Println("Listening on localhost:5000")
	err = http.ListenAndServe("localhost:5000", mux)
	if err != nil {
		log.Fatalf("Error: %+v", err)
	}
}

func handleIndexRoute(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
	<html lang="en">
	<body>
		<h1>Hello world</h1>
	</body>
	</html>`
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func handleSubmitSATForm(w http.ResponseWriter, r *http.Request) {
}

func handleGetRank(w http.ResponseWriter, r *http.Request) {
}

func handleUpdateSATScore(w http.ResponseWriter, r *http.Request) {
}

func handleDeleteRecord(w http.ResponseWriter, r *http.Request) {
}

func handleViewAllData(w http.ResponseWriter, r *http.Request) {
}
