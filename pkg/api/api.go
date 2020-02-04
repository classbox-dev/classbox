package api

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mkuznets/classbox/pkg/opts"
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
	App         *opts.App
	AWS         *opts.AWS
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
		r.Route("/signup", func(r chi.Router) {
			r.Get("/oauth", s.API.OAuthURL)
			r.Post("/create", s.API.CreateUser)
			r.Post("/install", s.API.InstallApp)
		})
		r.Route("/commits", func(r chi.Router) {
			r.Get("/{login}:{commitHash:[0-9a-z]+}", s.API.GetCommit)
		})
		r.Route("/tasks", func(r chi.Router) {
			r.Post("/{taskID:[0-9a-z-]+}", s.API.FinishTask)
			r.Post("/dequeue", s.API.DequeueTask)
			r.Post("/enqueue", s.API.EnqueueTask)
		})
		r.Get("/runs", s.API.GetRuns)
		r.Put("/runs", s.API.CreateRuns)
		r.Get("/meta", s.API.GetMeta)
		r.Put("/meta", s.API.UpdateMeta)
	})

	err := http.ListenAndServe(s.Addr, router)
	if err != nil {
		log.Printf("[WARN] server has terminated: %s", err)
	}
}
