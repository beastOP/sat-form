package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
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

// The following comment and the variable below is important as it tells the
// go compiler to embed the migrations folder into the binary.
//
//go:embed migrations/*.sql
var embedMigrations embed.FS

// I opted to only the standard library to implement the http server as it
// was a small project and I wanted to avoid adding additional dependencies.
//
// We start by opening a connection to the database. We then check if we can
// ping the database. If we can, we run the migrations. If we can't, we log
// the error and exit.
//
// We then create a new servemux and a file server to serve the static files.
//
// We then handle the routes.
//
// We then start the HTTP server.
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

	// Create the queries struct generated from the `sqlc generate` command
	// as it is needed to interact with the database.
	queries := database.New(db)

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static")) // Creating a file server to serve the static files.
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Handle the frontend routes.
	mux.HandleFunc("/", handleHtmlError(handleIndexRoute(queries)))
	mux.HandleFunc("/update-sat-score-form", handleHtmlError(handleUpdateSATScoreForm(queries)))

	apiMux := http.NewServeMux() // Creating `/api` group for api routes.
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// Handle the api routes.
	apiMux.HandleFunc("POST /submit-sat-score", handleApiError(handleSubmitSATForm(db, queries)))
	apiMux.HandleFunc("GET /search-by-name", handleApiError(handleSearchByName(queries)))
	apiMux.HandleFunc("POST /update-sat-score", handleApiError(handleUpdateSATScore(db, queries)))
	apiMux.HandleFunc("DELETE /delete-record", handleApiError(handleDeleteRecord(db, queries)))
	apiMux.HandleFunc("GET /view-all-data", handleApiError(handleViewAllData(queries)))

	fmt.Println("Listening on localhost:5000")
	err = http.ListenAndServe("localhost:5000", mux)
	if err != nil {
		log.Fatalf("Error: %+v", err)
	}
}

func renderJSON(w http.ResponseWriter, r *http.Request, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

func render(w http.ResponseWriter, r *http.Request, statusCode int, t templ.Component) error {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(statusCode)
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

func handleUpdateSATScoreForm(queries *database.Queries) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		name := r.URL.Query().Get("name")
		satScore, err := queries.GetSATScoreByName(r.Context(), name)
		if err != nil {
			return err
		}
		return render(w, r, http.StatusOK, templates.UpdateForm(satScore.Name, satScore.SatScore))
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
		if satScore < 0 || satScore > 100 {
			return errors.New("value out of range (0-100)")
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

func handleSearchByName(queries *database.Queries) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		name := r.URL.Query().Get("name")
		like := "%" + name + "%"
		satScores, err := queries.GetNameBySubstring(r.Context(), like)
		if err != nil {
			return err
		}
		return render(w, r, http.StatusOK, templates.Table(satScores))
	}
}

func handleUpdateSATScore(db *sql.DB, queries *database.Queries) func(w http.ResponseWriter, r *http.Request) error {
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
		if satScore < 0 || satScore > 100 {
			return errors.New("value out of range (0-100)")
		}
		_, err = queries.UpdateSATScore(r.Context(), database.UpdateSATScoreParams{
			Name:     r.FormValue("name"),
			SatScore: satScore,
			Passed:   satScore >= 30,
		})
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
		render(w, r, http.StatusOK, templates.FormWithTable(satScores))
		return nil
	}
}

func handleDeleteRecord(db *sql.DB, queries *database.Queries) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		name := r.URL.Query().Get("name")
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		queries = queries.WithTx(tx)
		err = queries.DeleteSATScore(r.Context(), name)
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

func handleViewAllData(queries *database.Queries) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		satScores, err := queries.GetSATScores(r.Context())
		if err != nil {
			return err
		}
		err = renderJSON(w, r, http.StatusOK, struct {
			SatScores []database.SatScore `json:"sat_scores"`
		}{SatScores: satScores})
		if err != nil {
			return err
		}
		return nil
	}
}
