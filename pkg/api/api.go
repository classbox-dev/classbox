package api

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// API is a collection of endpoints
type API struct {
	DB          *pgxpool.Pool
	OAuth       *oauth2.Config
	RandomState string
}

// Server is a
type Server struct {
	Addr string
	Port int
	API  API
}

// Start initialises the server
func (s *Server) Start() {
	router := chi.NewRouter()

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	router.Use(middleware.Timeout(60 * time.Second))

	router.Route("/", func(r chi.Router) {
		r.Get("/scoreboard", s.API.Scoreboard)
		r.Get("/signup", s.API.Signup)
	})

	err := http.ListenAndServe(s.Addr, router)
	if err != nil {
		log.Printf("[WARN] server has terminated: %s", err)
	}
}
