package main

import (
	"net/http"
	"sync"
)

type userData struct {
	username string
	password string
	email    string
}

type serviceData struct {

	userMap map[string]userData
	sessionMap map[string]string
	mtx sync.Mutex
}

func (data *serviceData) muxSetup() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/welcome/", http.HandlerFunc(data.welcomeHandler))	
	mux.Handle("/signup/", http.HandlerFunc(data.signUpHandler))
	mux.Handle("/user/signin/", http.HandlerFunc(data.signInHandler))
	mux.Handle("/user/account/", http.HandlerFunc(data.accountHandler))
	mux.Handle("/user/signout/", http.HandlerFunc(data.signOutHandler))
	
	return mux
}

func main() {
	data := &serviceData{}
	data.userMap = make(map[string]userData)
	data.sessionMap = make(map[string]string)
	mux := data.muxSetup()
	http.ListenAndServe(":8080", mux)
}
