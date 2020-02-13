package web

import (
	"github.com/mkuznets/classbox/pkg/api/client"
	"github.com/mkuznets/classbox/pkg/opts"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Web struct {
	API       *client.Client
	DocsURL   string
	WebURL    string
	Templates *Templates
}

type Server struct {
	Addr string
	Env  *opts.Env
	Port int
	Web  *Web
}

func (s *Server) Start() {
	log.Printf("[INFO] environment: %s", s.Env.Type)

	router := chi.NewRouter()

	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(10 * time.Second))

	router.Route("/", func(r chi.Router) {
		r.With(sessionAuth(s.Web.API.GetUser)).Get("/", s.Web.GetIndex)
		r.Get("/signin", s.Web.GetSignin)
		r.Get("/logout", s.Web.Logout)
		r.Get("/commit/{login}:{commitHash:[0-9a-z]+}", s.Web.GetCommit)
	})
	router.NotFound(s.Web.NotFound)

	err := http.ListenAndServe(s.Addr, router)
	if err != nil {
		log.Printf("[WARN] server has terminated: %s", err)
	}
}
