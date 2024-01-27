package handlers

import (
	"net/http"
)

func GraphCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func GetFriends(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetFriends"))
}

func DeleteFriend(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("DeleteFriend"))
}

func GetPendingFriend(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetPendingFriend"))
}

func GetFriendsExtended(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetFriendsExtendedGraph"))
}
