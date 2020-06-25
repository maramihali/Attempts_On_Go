package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
)

type userData struct {
	username string
	password string
	email    string
}

var (

	// Maps usernames to user's passwords
	usersMap = make(map[string]userData)
	// Maps sessions to usernames
	sessionMap = make(map[string]string)
)

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	names, ok := r.URL.Query()["name"]

	if !ok || len(names[0]) < 1 {
		http.Error(w, "400 Bad Request", 400)
		return
	}

	name := names[0]
	fmt.Fprintf(w, "<h1>Hello, %s!</h1>", name)
}

func signUpHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "signup_form.html")
	case "POST":
		signUpPostHandler(w, r)
	default:
		fmt.Fprintf(w, "Sent a %s request when GET or POST requests were expected.", r.Method)
	}
}

func signUpPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "Failed to parse the form: %v. ", err)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")

	if password != confirmPassword {
		fmt.Fprint(w, "Submitted passwords differ.")
		return
	}

	if _, taken := usersMap[username]; taken {
		fmt.Fprint(w, "Username already taken, please try again")
		return
	}

	usersMap[username] = userData{
		username: username, password: password, email: email}

	fmt.Fprintf(w, "Successfully signed up, %s! Please sign in", usersMap[username].username)
}

func signInHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		fmt.Fprintf(w, `
		<h1>Welcome to the signin form!</h1>
		<form method="POST" action="/user/signin/">
			<label>Username:</label><input name="username" type="text" value="" />
			<label>Password:</label><input name="password" type="text" value="" />
			<input type="submit" value="Sign In" />
		</form>
	`)
	case "POST":
		signInPostHandler(w, r)
	default:
		fmt.Fprintf(w, "Sent a %s request when GET or POST requests were expected.", r.Method)
	}

}

func signInPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "Failed to parse the form: %v. ", err)
		return
	}

	username := r.FormValue("username")

	if _, exists := usersMap[username]; !exists {
		fmt.Fprintf(w, "User doesn't exist, please sign up.")
		return
	}

	data := usersMap[username]

	if data.password != r.FormValue("password") {
		fmt.Fprintf(w, "Invalid Password.")
		return
	}

	token := randToken()
	sessionMap[token] = username
	log.Printf("token: %s, username: %s", token, sessionMap[token])
	userCookie := &http.Cookie{Name: "session", Value: token, Path: "/user"}
	http.SetCookie(w, userCookie)
	log.Printf("%s %s", userCookie.Name, userCookie.Value)
	fmt.Fprintf(w, "Successfully logged in. You can see your account at /account/")
}

func signOutHandler(w http.ResponseWriter, r *http.Request) {

	if userCookie, err := r.Cookie("session"); err == nil {
		if _, exists := sessionMap[userCookie.Value]; exists {
			delete(sessionMap, userCookie.Value)
			cookie := &http.Cookie{Name: "session", Value: userCookie.Value, Path: "/user", MaxAge: -1}
			http.SetCookie(w, cookie)
			fmt.Fprintf(w, "Successfully signed out!")
		}
		return
	}
	fmt.Fprintf(w, "You have already signed out.")
}

func accountHandler(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("session")
	if err != nil {
		fmt.Fprintf(w, "Please sign in/up")
		return
	}

	username, exists := sessionMap[cookie.Value]

	if !exists {
		fmt.Fprintf(w, "Invalid session token.")
		return
	}
	data := usersMap[username]
	fmt.Fprintf(w, "Your email address is: %s", data.email)
}

func randToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func main() {

	mux := http.NewServeMux()

	wh := http.HandlerFunc(welcomeHandler)
	suh := http.HandlerFunc(signUpHandler)
	sih := http.HandlerFunc(signInHandler)
	soh := http.HandlerFunc(signOutHandler)
	ah := http.HandlerFunc(accountHandler)

	mux.Handle("/welcome/", wh)
	mux.Handle("/signup/", suh)
	mux.Handle("/user/signin/", sih)
	mux.Handle("/user/account/", ah)
	mux.Handle("/user/signout/", soh)

	http.ListenAndServe(":8080", mux)
}
