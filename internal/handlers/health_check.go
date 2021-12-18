package handlers

import (
	"github.com/jeffreyrogers/backtalk/internal/models"
	"net/http"
)

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
