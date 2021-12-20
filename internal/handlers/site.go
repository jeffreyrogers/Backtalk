package handlers

import (
	"crypto/subtle"
	"embed"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/jeffreyrogers/backtalk/internal/crypto"
	"github.com/jeffreyrogers/backtalk/internal/models"
	"github.com/jeffreyrogers/backtalk/internal/sqlc"
	"github.com/jeffreyrogers/backtalk/resources"
)

var res embed.FS

func init() {
	res = resources.Res
}

func ShowLogin(w http.ResponseWriter, r *http.Request) {
	// FIXME: this is a hack. The admin will always have uid == 0, so this is fine,
	// but we should actually verify that the uid returned has admin permissions.
	if loggedIn(r) == 0 {
		http.Redirect(w, r, "/admin", 302)
		return
	}

	tpl, err := template.ParseFS(res, "login.html")
	if err != nil {
		log.Printf("Error loading template %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	data := map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
		"title":          "Backtalk Login",
	}

	if err := tpl.Execute(w, data); err != nil {
		return
	}
}

func ShowRegister(w http.ResponseWriter, r *http.Request) {
	// FIXME: this is a hack. The admin will always have uid == 0, so this is fine,
	// but we should actually verify that the uid returned has admin permissions.
	if loggedIn(r) == 0 {
		http.Redirect(w, r, "/admin", 302)
		return
	}

	tpl, err := template.ParseFS(res, "register.html")
	if err != nil {
		log.Printf("Error loading template %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	data := map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
		"title":          "Backtalk Register",
	}

	if err := tpl.Execute(w, data); err != nil {
		return
	}
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	// Can only register a user if there are no users in the user table.
	populated, err := models.Queries.UsersPopulated(models.Ctx)
	if err != nil {
		log.Printf("Error running UsersPopulated query: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if populated {
		log.Printf("Attempted to register after user already created")
		http.Redirect(w, r, "/", 302)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	salt := crypto.GenerateSalt()
	hash := crypto.Hash(password, salt)

	_, err = models.Queries.CreateAdminUser(models.Ctx, sqlc.CreateAdminUserParams{email, hash, salt})
	if err != nil {
		log.Printf("Error inserting admin user into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", 302)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	// get email and password
	email := r.FormValue("email")
	password := r.FormValue("password")

	log.Printf("Email: %s", email)
	log.Printf("Password: %s", password)

	// get associated user from db
	user, err := models.Queries.GetUser(models.Ctx, email)
	if err != nil {
		// TODO: set a cookie so that the form can alert the user of the problem
		log.Printf("Unable to get user: %v", err)
		http.Redirect(w, r, "/", 302)
		return
	}

	if !crypto.PasswordValid(user.Hash, password, user.Salt) {
		// TODO: set a cookie so that the form can alert the user of the problem
		log.Printf("Password does not match")
		http.Redirect(w, r, "/", 302)
		return
	}

	id, key := crypto.GenerateSessionKey()

	// store id in database
	err = models.Queries.CreateSession(models.Ctx, sqlc.CreateSessionParams{id, user.ID})
	if err != nil {
		log.Printf("Error inserting session into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO set expires and also refresh expires
	cookie := &http.Cookie{
		Name:     "sessionKey",
		Value:    key,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/admin", 302)
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

	sessionID, ok := crypto.SessionIDValid(sessionKeyCookie.Value)
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
