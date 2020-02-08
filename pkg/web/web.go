package web

import (
	"github.com/mkuznets/classbox/pkg/api/client"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Web struct {
	API       *client.Client
	Templates *Templates
}

type Server struct {
	Addr string
	Port int
	Web  *Web
}

func (s *Server) Start() {
	router := chi.NewRouter()

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	router.Use(middleware.Timeout(3 * time.Second))

	router.Route("/", func(r chi.Router) {
		r.Get("/", s.Web.GetIndex)
		r.Get("/commit/{login}:{commitHash:[0-9a-z]+}", s.Web.GetCommit)
	})
	router.NotFound(s.Web.NotFound)

	err := http.ListenAndServe(s.Addr, router)
	if err != nil {
		log.Printf("[WARN] server has terminated: %s", err)
	}
}
