package web

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/mkuznets/classbox/pkg/api/client"
	"net/http"
)

func sessionAuth(API func(r *http.Request) *client.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			api := API(r)
			user, err := api.GetUser(r.Context())
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), "User", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func validateProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if chi.URLParam(r, "project") != "stdlib" {
			http.Redirect(w, r, "/stdlib", http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}
