package api

import (
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v4"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/mkuznets/classbox/pkg/api/models"
	"net/http"
)

func (api *API) GetUser(w http.ResponseWriter, r *http.Request) {
	session := r.Header.Get("X-Session")
	if session == "" {
		E.SendError(w, r, nil, http.StatusBadRequest, "`X-Session` header expected")
		return
	}
	var user models.User
	err := api.DB.QueryRow(r.Context(), `
	SELECT u.login, u.repository_name
	FROM users as u JOIN sessions as s ON (s.user_id=u.id)
	WHERE session=$1 LIMIT 1
	`, session).Scan(&user.Login, &user.Repo)
	switch {
	case err == pgx.ErrNoRows:
		E.SendError(w, r, nil, http.StatusUnauthorized, "user not authenticated")
		return
	case err != nil:
		E.Handle(w, r, err)
		return
	}
	render.JSON(w, r, &user)
}
