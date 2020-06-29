package main

import (
	"html/template"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
)

// TODO(maramihali22): use template/html

func (data *serviceData) welcomeHandler(w http.ResponseWriter, r *http.Request){
	names, ok := r.URL.Query()["name"]
	if !ok || len(names[0]) < 1 {
		http.Error(w, "400 Bad Request", http.StatusBadRequest)
		return
	}
	t, err := template.New("welcome").Parse(`{{define "T"}} Hello, {{.}}!{{end}}`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// Assume the first query parameter correspond to the user's name
	name := names[0]
	err = t.ExecuteTemplate(w, "T", name)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (data *serviceData) signUpHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		data.singUpGetHandlers(w, r)
	case "POST":
		data.signUpPostHandler(w, r)
	default:
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (data *serviceData) singUpGetHandlers(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("forms.html"))
	err := t.Execute(w,struct{AccountExists bool}{false})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)	
	}
}

func (data *serviceData) signUpPostHandler(w http.ResponseWriter, r *http.Request) {	
	data.mtx.Lock()
	defer data.mtx.Unlock()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse the form: %v. ", http.StatusInternalServerError)
		return
	}

	u := userData{
        username: r.FormValue("username"),
        email: r.FormValue("email"),
        password: r.FormValue("password"),
	}

	confirmPassword := r.FormValue("confirmPassword")

	if u.password != confirmPassword {
		http.Error(w, "Submitted passwords differ.", http.StatusBadRequest)
		return
	}

	if _, taken := data.userMap[u.username]; taken {
		http.Error(w, "Username already taken, please try again", http.StatusBadRequest)
		return
	}

	data.userMap[u.username] = u

	fmt.Fprintf(w, "Successfully signed up, %s! Please sign in", data.userMap[u.username].username)
}

func (data *serviceData) signInHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		data.signInGetHandler(w, r)
	case "POST":
		data.signInPostHandler(w, r)
	default:
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
	}

}

func (data *serviceData) signInGetHandler(w http.ResponseWriter, r *http.Request) {
	
	t := template.Must(template.ParseFiles("forms.html"))
	err := t.Execute(w,struct{AccountExists bool}{true})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)	
	}
}

func (data *serviceData) signInPostHandler(w http.ResponseWriter, r *http.Request) {
	data.mtx.Lock()
	defer data.mtx.Unlock()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse the form: %v. ", http.StatusInternalServerError)
		return
	}

	username := r.FormValue("username")

	if _, exists := data.userMap[username]; !exists {
		http.Redirect(w, r, "%2e%2e%2fsignup", http.StatusSeeOther)
		return
	}

	userData := data.userMap[username]

	if userData.password != r.FormValue("password") {
		return
	}

	token, _ := rand.Int(rand.Reader, big.NewInt(1000))
	keyToken := token.String()
	
	if _, exists := data.sessionMap[keyToken]; exists {
		panic("Cryptography failed us!")
	}

	data.sessionMap[keyToken] = username
	log.Printf("token: %s, username: %s", keyToken, data.sessionMap[keyToken])
	sessionCookie := &http.Cookie{Name: "session", Value: keyToken, Path: "/user"}
	http.SetCookie(w, sessionCookie)
	log.Printf("%s %s", sessionCookie.Name, sessionCookie.Value)
	fmt.Fprintf(w, "Successfully logged in. You can see your account at /account/")
}

func (data *serviceData) signOutHandler(w http.ResponseWriter, r *http.Request) {
	data.mtx.Lock()
	defer data.mtx.Unlock()

	sessionCookie, err := r.Cookie("session")

	if err != nil {
		io.WriteString(w, "Error: cookie not found.")
		return
	}

	if _, exists := data.sessionMap[sessionCookie.Value]; exists {
			fmt.Fprintf(w, "You have already signed out.")
			return
	}

	delete(data.sessionMap, sessionCookie.Value)
	cookie := &http.Cookie{Name: "session", Value: sessionCookie.Value, Path: "/user", MaxAge: -1}
	http.SetCookie(w, cookie)
	fmt.Fprintf(w, "Successfully signed out!")
}

func (data *serviceData) accountHandler(w http.ResponseWriter, r *http.Request) {
	data.mtx.Lock()
	defer data.mtx.Unlock()

	cookie, err := r.Cookie("session")
	if err != nil {
		fmt.Fprintf(w, "Please sign in/up")
		return
	}

	username, exists := data.sessionMap[cookie.Value]

	if !exists {
		http.Error(w, "401 Status Unauthorized", http.StatusUnauthorized)
		return
	}
	userData := data.userMap[username]
	fmt.Fprintf(w, "Your email address is: %s", userData.email)
}
