package handlers

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jeffreyrogers/backtalk/internal/globals"
	"github.com/jeffreyrogers/backtalk/internal/sqlc"
)

func CreateComment(w http.ResponseWriter, r *http.Request) {
	// get slug from
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		w.WriteHeader(400)
		log.Println("No slug provided to CreateComment")
		return
	}

	author := r.FormValue("author")
	if author == "" {
		w.WriteHeader(400)
		log.Println("No author provided to CreateComment")
		return
	}

	comment := r.FormValue("comment")
	if comment == "" {
		w.WriteHeader(400)
		log.Println("No comment provided to CreateComment")
		return
	}

	_, err := globals.Queries.CreateComment(globals.Ctx, sqlc.CreateCommentParams{slug, author, comment})
	if err != nil {
		w.WriteHeader(503)
		log.Println("Could not create comment")
		return
	}

	w.WriteHeader(200)
	log.Println("Created comment")
}

func GetComments(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("Returned comments"))
	log.Println("Returned comments")
}

func EditComment(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("Edited comment"))
	log.Println("Edited comment")
}

func DeleteComment(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("Deleted comment"))
	log.Println("Deleted comment")
}
