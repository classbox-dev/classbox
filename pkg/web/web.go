package web

import (
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mkuznets/classbox/pkg/api/client"
	"github.com/mkuznets/classbox/pkg/opts"
	"github.com/rakyll/statik/fs"
	"log"
	"net/http"
	"time"
)

type Server struct {
	Addr   string
	Sentry *opts.Sentry
	Env    *opts.Env
	Port   int
	Web    *Web
}

func (s *Server) Start() {
	log.Printf("[INFO] environment: %s", s.Env.Type)

	staticFs, err := fs.New()
	if err != nil {
		log.Fatalf("could not initialise statik fs: %v", err)
	}
	staticServer := http.FileServer(staticFs)

	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(10 * time.Second))

	if s.Sentry.Init(s.Env.Type, "web") {
		sentryMiddleware := sentryhttp.New(sentryhttp.Options{
			Repanic: true,
			Timeout: 10 * time.Second,
		})
		router.Use(sentryMiddleware.Handle)
	}

	router.Route("/", func(r chi.Router) {
		r.Mount("/.static", staticServer)
		r.Mount(`/{:*\.(png|svg|ico|webmanifest)}`, staticServer)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/stdlib", http.StatusMovedPermanently)
			return
		})
		router.With(validateProject).Route("/{project:[0-9a-z]+}", func(r chi.Router) {
			r.With(sessionAuth(s.Web.API)).Group(func(r chi.Router) {
				r.Get("/", s.Web.GetIndex)
				r.Get("/scoreboard", s.Web.GetScoreboard)
				r.Get("/commit/{login}:{commitHash:[0-9a-z]+}", s.Web.GetCommit)
				r.Get("/quickstart", s.Web.GetQuickstart)
				r.Get("/prerequisites", s.Web.GetPrerequisites)
			})
			r.Get("/signin", s.Web.GetSignin)
			r.Get("/logout", s.Web.Logout)
		})
	})

	router.NotFound(s.Web.NotFound)

	err = http.ListenAndServe(s.Addr, router)
	if err != nil {
		log.Printf("[WARN] server has terminated: %s", err)
	}
}

type Web struct {
	DocsURL   string
	ApiURL    string
	WebURL    string
	Templates *Templates
}

func (web *Web) API(r *http.Request) *client.Client {
	cl := client.New(web.ApiURL)
	cookie, err := r.Cookie("session")
	if err != nil {
		return cl
	}
	cl.SessionAuth(cookie.Value)
	return cl
}
