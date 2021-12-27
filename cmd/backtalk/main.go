package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
	"github.com/jeffreyrogers/backtalk/internal/globals"
	"github.com/jeffreyrogers/backtalk/internal/handlers"
	"github.com/jeffreyrogers/backtalk/internal/sqlc"

	_ "github.com/lib/pq"
)

// `main()` is responsible for two things:
//
//     1. Setting up our connection to the database so that other packages can query it
//	   2. Setting up the routing
//
// All other functionality is handled in the route handlers or by go code generated by sqlc that interacts with the database.
func main() {
	globals.Ctx = context.Background()

	ticker := time.NewTicker(time.Hour)
	done := make(chan bool)

	authString := os.Getenv("BACKTALK_AUTH_KEY")

	var err error
	globals.AuthKey, err = base64.StdEncoding.DecodeString(authString)
	if err != nil {
		log.Fatal(err)
	}

	// most of the connection string arguments are handled via environment variables (e.g. PGHOST, PGPASS, etc.)
	globals.DB, err = sql.Open("postgres", "application_name=backtalk idle_in_transaction_session_timeout=10000 statement_timeout=10000")
	if err != nil {
		log.Fatal(err)
	}

	// I have no idea if this is an appropriate number or not. Will have to benchmark to check, but doesn't matter for now.
	globals.DB.SetMaxOpenConns(20)
	globals.Queries = sqlc.New(globals.DB)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// run cleanup
				err := globals.Queries.DeleteOldSessions(globals.Ctx)
				if err != nil {
					log.Println("Failed to delete expired sessions: %v", err)
				}
				log.Println("Checking for expired sessions")
			}
		}
	}()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", handlers.ShowLogin)
	r.Post("/login", handlers.LoginUser)
	r.Get("/register", handlers.ShowRegister)
	r.Post("/register", handlers.RegisterUser)
	r.Get("/health", handlers.HealthCheck)

	// REST API
	r.Route("/comments", func(r chi.Router) {
		r.Post("/{slug}", handlers.CreateComment)
		r.Get("/{slug}", handlers.GetComments)
		r.Delete("/{slug}/{id}", handlers.DeleteComment)
		r.Put("/{slug}/{id}", handlers.EditComment)
	})

	r.Mount("/admin", adminRouter())
	CSRF := csrf.Protect(globals.AuthKey)

	log.Println("Starting server on port 8000")
	http.ListenAndServe(":8000", CSRF(r))

	ticker.Stop()
	done <- true
	log.Println("Ticker stopped")
}

func adminRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(handlers.AdminOnly)
	r.Get("/", handlers.AdminHome)
	return r
}
