package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sat-form/database"
	"sat-form/templates"
	"strconv"

	"github.com/a-h/templ"
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

	queries := database.New(db)

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/", handleHtmlError(handleIndexRoute(queries)))

	apiMux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	apiMux.HandleFunc("/submit-sat-score", handleApiError(handleSubmitSATForm(db, queries)))
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

func renderJSON(w http.ResponseWriter, r *http.Request, statusCode int, data any) error {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

func render(w http.ResponseWriter, r *http.Request, statusCode int, t templ.Component) error {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "text/html")
	return t.Render(r.Context(), w)
}

func handleHtmlError(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			log.Printf("Error: %+v", err)
			render(w, r, http.StatusInternalServerError, templates.Error(err.Error()))
		}
	}
}

func handleApiError(f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			log.Printf("Error: %+v", err)
			renderJSON(w, r, http.StatusInternalServerError, err.Error())
		}
	}
}

func handleIndexRoute(queries *database.Queries) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		satScores, err := queries.GetSATScores(r.Context())
		if err != nil {
			return err
		}
		index := templates.Index(satScores)
		err = render(w, r, http.StatusOK, index)
		if err != nil {
			return err
		}
		return nil
	}
}

func handleSubmitSATForm(db *sql.DB, queries *database.Queries) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		queries = queries.WithTx(tx)
		satScore, err := strconv.ParseInt(r.FormValue("sat_score"), 10, 64)
		if err != nil {
			return err
		}
		newSatScore := database.InsertSATScoreParams{
			Name:     r.FormValue("name"),
			Address:  r.FormValue("address"),
			City:     r.FormValue("city"),
			Country:  r.FormValue("country"),
			Pincode:  r.FormValue("pincode"),
			SatScore: satScore,
			Passed:   satScore >= 30,
		}
		_, err = queries.InsertSATScore(r.Context(), newSatScore)
		if err != nil {
			return err
		}
		err = queries.UpdateSATScoreRanks(r.Context())
		if err != nil {
			log.Printf("Error: %+v", err)
			tx.Rollback()
			return err
		}
		satScores, err := queries.GetSATScores(r.Context())
		if err != nil {
			log.Printf("Error: %+v", err)
			tx.Rollback()
			return err
		}
		tx.Commit()
		render(w, r, http.StatusOK, templates.Table(satScores))
		return nil
	}
}

func handleGetRank(w http.ResponseWriter, r *http.Request) {
}

func handleUpdateSATScore(w http.ResponseWriter, r *http.Request) {
}

func handleDeleteRecord(w http.ResponseWriter, r *http.Request) {
}

func handleViewAllData(w http.ResponseWriter, r *http.Request) {
}
