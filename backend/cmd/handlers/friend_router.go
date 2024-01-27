package handlers

import (
	"net/http"
)

func GraphCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func GetGraph(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetGraph"))
}

func GetFriendsExtendedGraph(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetFriendsExtendedGraph"))
}
