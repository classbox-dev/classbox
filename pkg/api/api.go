package api

import (
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mkuznets/classbox/pkg/opts"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"time"
)

// API is a collection of endpoints
type API struct {
	DB          *pgxpool.Pool
	OAuth       *oauth2.Config
	App         *opts.App
	AWS         *opts.AWS
	Jwt         *opts.JwtServer
	RandomState string
	WebUrl      string
}

// Server is a
type Server struct {
	Addr   string
	Sentry *opts.Sentry
	Env    *opts.Env
	Port   int
	API    API
}

// Start initialises the server
func (s *Server) Start() {
	log.Printf("[INFO] environment: %s", s.Env.Type)

	router := chi.NewRouter()
	router.Use(middleware.Timeout(30 * time.Second))
	router.Use(middleware.Recoverer)

	if s.Sentry.Init(s.Env.Type, "api") {
		sentryMiddleware := sentryhttp.New(sentryhttp.Options{
			Repanic: true,
			Timeout: 10 * time.Second,
		})
		router.Use(sentryMiddleware.Handle)
	}

	router.Route("/", func(r chi.Router) {

		// web endpoints
		r.Get("/stats", s.API.GetStats)
		r.Route("/signin", func(r chi.Router) {
			r.Get("/oauth", s.API.OAuthURL)
			r.Post("/create", s.API.CreateUser)
			r.Post("/install", s.API.InstallApp)
		})
		r.Get("/commits/{login}:{commitHash:[0-9a-z]+}", s.API.GetCommit)
		r.Get("/tests", s.API.GetTests)
		r.Get("/user", s.API.GetUser)

		// webhook endpoint
		r.With(hookValidator(s.API.App.HookSecret)).Post("/tasks/enqueue", s.API.EnqueueTask)

		// private runner's endpoints
		r.With(jwtValidator(s.API.Jwt.Key)).Route("/", func(r chi.Router) {
			r.Route("/course", func(r chi.Router) {
				r.Get("/", s.API.GetCourse)
				r.Put("/", s.API.UpdateCourse)
			})
			r.Put("/tests", s.API.UpdateTests)
			r.Route("/runs", func(r chi.Router) {
				r.Get("/", s.API.GetRuns)
				r.Put("/", s.API.CreateRuns)
				r.Get("/baselines", s.API.GetBaselines)
			})
			r.Route("/tasks", func(r chi.Router) {
				r.Post("/{taskID:[0-9a-z-]+}", s.API.FinishTask)
				r.Post("/dequeue", s.API.DequeueTask)
			})
		})
	})

	if err := http.ListenAndServe(s.Addr, router); err != nil {
		log.Printf("[WARN] server has terminated: %s", err)
	}
}
