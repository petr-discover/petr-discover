package handlers

import (
	"net/http"
)

func FriendCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func GetGraph(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetFriends"))
}
