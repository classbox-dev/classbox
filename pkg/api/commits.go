package api

import (
	"database/sql"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/pkg/errors"
	"net/http"
)

type commitResponse struct {
	Login  string           `json:"login"`
	Repo   string           `json:"repository"`
	Commit string           `json:"commit"`
	Status string           `json:"status"`
	Checks []*checkResponse `json:"checks"`
}

type checkResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Output string `json:"output"`
	Cached bool   `json:"is_cached"`
}

func (api *API) Commit(w http.ResponseWriter, r *http.Request) {

	commitHash := chi.URLParam(r, "commitHash")
	login := chi.URLParam(r, "login")

	resp := commitResponse{Commit: commitHash}
	resp.Checks = []*checkResponse{}

	var commitID uint64

	err := api.DB.QueryRow(r.Context(), `
	SELECT c.id, u.login, u.repository_name, UPPER(t.status::text)
	FROM
		commits AS c
		JOIN users AS u ON(c.user_id=u.id)
		JOIN tasks AS t ON(c.user_id=u.id)
	WHERE
		c.commit=$1 AND u.login=$2
	;`, commitHash, login).Scan(&commitID, &resp.Login, &resp.Repo, &resp.Status)

	switch {
	case err == sql.ErrNoRows:
		E.Render(w, r, E.NotFound(errors.New("unknown commit")))
		return
	case err != nil:
		E.Render(w, r, E.Internal(errors.Wrap(err, "query error")))
		return
	}

	rows, err := api.DB.Query(r.Context(), `
	SELECT name, UPPER(status::text), output, is_cached
	FROM checks
	WHERE commit_id=$1
	ORDER BY test_id, id
	;`, commitID)
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "query error")))
		return
	}

	defer rows.Close()

	// Iterate through the result set
	for rows.Next() {
		var check checkResponse
		err = rows.Scan(&check.Name, &check.Status, &check.Output, &check.Cached)
		if err != nil {
			E.Render(w, r, E.Internal(errors.Wrap(err, "query error")))
			return
		}
		resp.Checks = append(resp.Checks, &check)
	}

	// Any errors encountered by rows.Next or rows.Scan will be returned here
	if rows.Err() != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "query error")))
		return
	}

	render.JSON(w, r, &resp)
}
