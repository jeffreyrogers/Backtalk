package handlers

import (
	"crypto/subtle"
	"github.com/jeffreyrogers/backtalk/internal/models"
	"github.com/jeffreyrogers/backtalk/internal/sessions"

	"net/http"
)

// TODO: make this a login page and redirect to admin page if logged in
func Home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

func AdminHome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello admin"))
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	comments, err := models.Queries.GetComments(models.Ctx, "test-slug")
	if err != nil {
		w.WriteHeader(503)
		w.Write([]byte("503 Service Unavailable"))
		return
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
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !loggedIn(r) {
			http.Redirect(w, r, "/", 302)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggedIn(r *http.Request) bool {
	sessionKeyCookie, err := r.Cookie("sessionKey")
	if err != nil {
		return false
	}

	sessionID, ok := sessions.SessionIDValid(sessionKeyCookie.Value)
	if !ok {
		return false
	}

	session, err := models.Queries.GetSession(models.Ctx, sessionID)
	if err != nil {
		return false
	}

	if subtle.ConstantTimeCompare([]byte(session.SessionID), []byte(sessionID)) == 1 {
		return true
	} else {
		return false
	}
}
