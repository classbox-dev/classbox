package api

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v4"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/db"
	"github.com/pkg/errors"
)

func (api *API) GetCommit(w http.ResponseWriter, r *http.Request) {

	commitHash := chi.URLParam(r, "commitHash")
	login := chi.URLParam(r, "login")

	var (
		commitID uint64
		resp     models.Commit
	)

	err := api.DB.QueryRow(r.Context(), `
	SELECT c.id, c.commit, u.login, u.repository_name, UPPER(t.status::text)
	FROM commits AS c JOIN users AS u ON(c.user_id=u.id) JOIN tasks AS t ON(c.id=t.commit_id)
	WHERE c.commit=$1 AND u.login=$2
	LIMIT 1
	`, commitHash, login).Scan(&commitID, &resp.Commit, &resp.Login, &resp.Repo, &resp.Status)

	switch {
	case err == pgx.ErrNoRows:
		e := fmt.Errorf("unknown commit: %s:%s", login, commitHash)
		E.SendError(w, r, e, http.StatusNotFound, e.Error())
		return
	case err != nil:
		E.Handle(w, r, err)
		return
	}

	rows, err := api.DB.Query(r.Context(), `
	SELECT name, UPPER(status::text), output FROM checks WHERE commit_id=$1 ORDER BY test_id NULLS FIRST, id
	`, commitID)
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	resp.Checks = make([]*models.Stage, 0)

	err = db.IterRows(rows, func(rows pgx.Rows) error {
		var check models.Stage
		err = rows.Scan(&check.Name, &check.Status, &check.Output)
		if err != nil {
			return errors.WithStack(err)
		}
		resp.Checks = append(resp.Checks, &check)
		return nil
	})

	render.JSON(w, r, &resp)
}
