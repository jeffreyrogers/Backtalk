package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jeffreyrogers/backtalk/internal/crypto"
	"github.com/jeffreyrogers/backtalk/internal/csrf"
	"github.com/jeffreyrogers/backtalk/internal/globals"
	"github.com/jeffreyrogers/backtalk/internal/sqlc"
	"github.com/jeffreyrogers/backtalk/resources"
)

var res embed.FS

func init() {
	res = resources.Res
}

func ShowLogin(w http.ResponseWriter, r *http.Request) {
	isAdmin, _ := loggedIn(r)
	if isAdmin {
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
		csrf.CSRFTag: csrf.CSRFField(r),
		"title":      "Backtalk Login",
	}

	if err := tpl.Execute(w, data); err != nil {
		return
	}
}

func ShowRegister(w http.ResponseWriter, r *http.Request) {
	isAdmin, _ := loggedIn(r)
	if isAdmin {
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
		csrf.CSRFTag: csrf.CSRFField(r),
		"title":      "Backtalk Register",
	}

	if err := tpl.Execute(w, data); err != nil {
		return
	}
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	// Can only register a user if there are no users in the user table.
	populated, err := globals.Queries.UsersPopulated(globals.Ctx)
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

	_, err = globals.Queries.CreateAdminUser(globals.Ctx, sqlc.CreateAdminUserParams{email, hash, salt})
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

	// get associated user from db
	user, err := globals.Queries.GetUser(globals.Ctx, email)
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
	err = globals.Queries.CreateSession(globals.Ctx, sqlc.CreateSessionParams{id, user.ID})
	if err != nil {
		log.Printf("Error inserting session into database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	secondsToExpiration := 60 * 60 * 24 * 30

	cookie := &http.Cookie{
		Name:     "sessionKey",
		Value:    key,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		MaxAge:   secondsToExpiration,
		Expires:  time.Now().Add(time.Duration(secondsToExpiration) * time.Second),
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/admin", 302)
}

func AdminHome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello admin"))
}

func ShowIP(w http.ResponseWriter, r *http.Request) {
	ip, err := getIP(r)
	if err != nil {
		w.WriteHeader(503)
		w.Write([]byte("Unable to Get IP Address"))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(ip))
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	comments, err := globals.Queries.GetComments(globals.Ctx, "test-slug")
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
		isAdmin, _ := loggedIn(r)
		if !isAdmin {
			http.Redirect(w, r, "/", 302)
			return
		}

		updateCookieExpiration(w, r)
		next.ServeHTTP(w, r)
	})
}

func updateCookieExpiration(w http.ResponseWriter, r *http.Request) {
	// we can ignore the error because just prior to calling this we successfully got the cookie in the loggedIn() function.
	c, _ := r.Cookie("sessionKey")
	secondsToExpiration := 60 * 60 * 24 * 30
	c.MaxAge = secondsToExpiration
	c.Expires = time.Now().Add(time.Duration(secondsToExpiration) * time.Second)
	c.Path = "/"
	c.Secure = true
	c.HttpOnly = true
	http.SetCookie(w, c)
}

// returns uid of logged in user.
// if user is not logged in return -1
// also return true if user has admin permissions, false otherwise
func loggedIn(r *http.Request) (bool, int32) {
	sessionKeyCookie, err := r.Cookie("sessionKey")
	if err != nil {
		return false, -1
	}

	sessionID, ok := crypto.SessionIDValid(sessionKeyCookie.Value)
	if !ok {
		log.Println("Invalid session id")
		return false, -1
	}

	user, err := globals.Queries.GetUserFromSession(globals.Ctx, sessionID)
	if err != nil {
		log.Printf("Could not get user from session: %v", err)
		return false, -1
	}

	err = globals.Queries.UpdateSessionLastSeen(globals.Ctx, sessionID)
	if err != nil {
		log.Printf("Could not update last seen time for user with id %d", user.ID)
	}

	return user.IsAdmin, user.ID
}

func getIP(r *http.Request) (string, error) {
	ips := r.Header.Get("X-FORWARDED-FOR")
	if ips == "" {
		return "", fmt.Errorf("No valid IP Found")
	}

	clientIP := strings.Split(ips, ",")[0]
	parsedIP := net.ParseIP(clientIP)
	if parsedIP != nil {
		return clientIP, nil
	}

	return "", fmt.Errorf("No valid IP Found")
}
