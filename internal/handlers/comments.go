package handlers

import (
	"log"
	"net/http"
)

func CreateComment(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("Created Comment"))
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
