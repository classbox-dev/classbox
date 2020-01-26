package api

import (
	"database/sql"
	"encoding/json"
	"github.com/go-chi/render"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/mkuznets/classbox/pkg/github"
	"github.com/pkg/errors"
	"net/http"
)

func (api *API) SubmissionHook(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("X-GitHub-Event") != "check_suite" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data := github.CheckSuiteEvent{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		E.Render(w, r, E.BadRequest(errors.Wrap(err, "invalid input")))
		return
	}

	var userID uint64
	err = api.DB.QueryRow(r.Context(), `
		SELECT "id" FROM "users" WHERE "github_id"=$1 AND "repository_id"=$2
	`, data.Sender.ID, data.Repo.ID).Scan(&userID)

	switch {
	case err == sql.ErrNoRows:
		E.Render(w, r, E.BadRequest(errors.Wrapf(err, "user not found: %s (id=%d)", data.Sender.Login, data.Sender.ID)))
		return
	case err != nil:
		E.Render(w, r, E.Internal(errors.Wrap(err, "query error")))
		return
	}

	appToken, err := api.App.Token()
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "could not get app token")))
		return
	}
	gh := github.New(appToken)
	err = gh.AuthAsInstallation(r.Context(), data.Inst.ID)
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "could not auth as installation")))
		return
	}

	checkRun, err := gh.CreateCheckRun(r.Context(), data.Repo.Owner.Login, data.Repo.Name, data.CheckSuite.Head)
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "could not create a check run")))
		return
	}

	tx, err := api.DB.Begin(r.Context())
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "could not start transaction")))
		return
	}
	//noinspection GoUnhandledErrorResult
	defer tx.Rollback(r.Context())

	var commitID uint64
	err = tx.QueryRow(r.Context(), `
		INSERT INTO commits ("user_id", "commit", "check_run_id")
		VALUES ($1, $2, $3)
		RETURNING "id"
		;
	`, userID, data.CheckSuite.Head, checkRun.ID).Scan(&commitID)
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "query error")))
		return
	}

	_, err = tx.Exec(r.Context(), `INSERT INTO "tasks" ("commit_id") VALUES ($1);`, commitID)
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "query error")))
		return
	}

	err = tx.Commit(r.Context())
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "could not commit transaction")))
		return
	}

	render.JSON(w, r, map[string]string{"url": "OK"})
}
