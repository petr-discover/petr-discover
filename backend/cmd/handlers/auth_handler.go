package handlers

import (
	"net/http"
)

func AuthCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("CreateUser"))
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Login"))
}

func Logout(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Logout"))
}
