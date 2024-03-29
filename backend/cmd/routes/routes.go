package routes

import (
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/petr-discover/cmd/handlers"
)

func NewRouter(port string) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		// AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		// ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		// MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	authRouter(r)
	userRouter(r)
	friendRouter(r)

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
		r.Post("/", handlers.CreateUserCard)
		r.Get("/", handlers.GetUser)
		r.Put("/", handlers.UpdateUser)
		r.Post("/friend", handlers.AddFriend)
	})
}

func friendRouter(r *chi.Mux) {
	r.Route("/api/v1/friends", func(r chi.Router) {
		r.Use(handlers.FriendCtx)
		r.Get("/pending", handlers.GetPendingFriend)
		r.Delete("/", handlers.DeleteFriend)
		r.Get("/", handlers.GetGraph)
	})
}
