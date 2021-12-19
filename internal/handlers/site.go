package handlers

import (
	"crypto/subtle"
	"github.com/jeffreyrogers/backtalk/internal/models"
	"github.com/jeffreyrogers/backtalk/internal/sessions"

	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	// FIXME: this is a hack. The admin will always have uid == 0, so this is fine,
	// but we should actually verify that the uid returned has admin permissions.
	if loggedIn(r) == 0 {
		http.Redirect(w, r, "/admin", 302)
		return
	}
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
		// FIXME: this is a hack. The admin will always have uid == 0, so this is fine,
		// but we should actually verify that the uid returned has admin permissions.
		if loggedIn(r) != 0 {
			http.Redirect(w, r, "/", 302)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// returns uid of logged in user, -1 if not logged in
func loggedIn(r *http.Request) int32 {
	sessionKeyCookie, err := r.Cookie("sessionKey")
	if err != nil {
		return -1
	}

	sessionID, ok := sessions.SessionIDValid(sessionKeyCookie.Value)
	if !ok {
		return -1
	}

	session, err := models.Queries.GetSession(models.Ctx, sessionID)
	if err != nil {
		return -1
	}

	if subtle.ConstantTimeCompare([]byte(session.SessionID), []byte(sessionID)) == 1 {
		return session.Uid
	} else {
		return -1
	}
}
