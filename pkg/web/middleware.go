package web

import (
	"context"
	"github.com/mkuznets/classbox/pkg/api/models"
	"net/http"
)

func sessionAuth(UserBySession func(ctx context.Context, session string) (*models.User, error)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := r.Cookie("session")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			user, err := UserBySession(r.Context(), session.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), "User", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
