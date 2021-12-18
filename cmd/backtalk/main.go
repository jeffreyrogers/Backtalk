package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jeffreyrogers/backtalk/internal/db"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()

	// most of the connection string arguments are handled via environment variables (e.g. PGHOST, PGPASS, etc.)
	pg, err := sql.Open("postgres", "application_name=backtalk idle_in_transaction_session_timeout=10000 statement_timeout=10000")
	if err != nil {
		log.Fatal(err)
	}

	// I have no idea if this is an appropriate number or not. Will have to benchmark to check, but doesn't matter for now.
	pg.SetMaxOpenConns(20)

	queries := db.New(pg)

	comments, err := queries.GetComments(ctx, "test-slug")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(comments)

	comments, err = queries.GetComments(ctx, "non-existent-slug")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(comments)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		comments, err := queries.GetComments(ctx, "test-slug")
		if err != nil {
			log.Fatal(err)
		}

		if len(comments) < 2 {
			w.WriteHeader(503)
			w.Write([]byte("503 Service Unavailable"))
			return
		}

		comment1 := comments[0]
		comment2 := comments[1]

		if comment1.Author != "John Doe" || comment1.Content != "Lorem ipsum dolor sit amet" {
			w.WriteHeader(503)
			w.Write([]byte("503 Service Unavailable"))
			return
		}
		if comment2.Author != "Jane Doe" || comment2.Content != "It was the best of times, it was the worst of times" {
			w.WriteHeader(503)
			w.Write([]byte("503 Service Unavailable"))
			return
		}

		w.WriteHeader(200)
		w.Write([]byte("200 OK"))
	})

	http.ListenAndServe(":8000", r)
}
