package routes

import (
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/petr-discover/cmd/handlers"
)

func NewRouter(port string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	authRouter(r)
	userRouter(r)
	friendRouter(r)
	graphRouter(r)

	log.Println("Server is running on port ", port)

	return r
}

func authRouter(r *chi.Mux) {
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Use(handlers.AuthCtx)
		r.Post("/register", handlers.CreateUser)
		r.Post("/login", handlers.Login)
		r.Post("/logout", handlers.Logout)
		r.Get("/google/login", handlers.GoogleAuth)
		r.Get("/google/callback", handlers.GoogleCallback)
	})
}

func userRouter(r *chi.Mux) {
	r.Route("/api/v1/user", func(r chi.Router) {
		r.Use(handlers.UserCtx)
		r.Get("/{id}", handlers.GetUser)
		r.Put("/{id}", handlers.UpdateUser)
	})
}

func friendRouter(r *chi.Mux) {
	r.Route("/api/v1/friends", func(r chi.Router) {
		r.Use(handlers.FriendCtx)
		r.Get("/", handlers.GetFriends)
		r.Post("/{id}", handlers.CreateFriend)
		r.Delete("/{id}", handlers.DeleteFriend)
	})
}

func graphRouter(r *chi.Mux) {
	r.Route("/api/v1/graph", func(r chi.Router) {
		r.Use(handlers.GraphCtx)
		r.Get("/", handlers.GetGraph)
		r.Get("/extended", handlers.GetFriendsExtendedGraph)
	})
}
