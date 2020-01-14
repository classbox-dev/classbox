package api

import (
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// API is a collection of endpoints
type API struct {
	DB *pgxpool.Pool
}

// Server is a
type Server struct {
	Port int
	API  API
}

// Start initialises the server
func (s *Server) Start() {
	router := chi.NewRouter()

	fmt.Println()

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	router.Use(middleware.Timeout(60 * time.Second))

	router.Route("/api", func(r chi.Router) {
		r.Get("/scoreboard", s.API.Scoreboard)
	})

	addr := fmt.Sprintf("0.0.0.0:%d", s.Port)
	err := http.ListenAndServe(addr, router)
	if err != nil {
		log.Printf("[WARN] server has terminated: %s", err)
	}
}
