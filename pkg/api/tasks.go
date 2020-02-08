package api

import (
	"bytes"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/db"
	"github.com/mkuznets/classbox/pkg/github"
	"github.com/mkuznets/classbox/pkg/s3"
	"github.com/mkuznets/classbox/pkg/utils"
	"github.com/pkg/errors"
	"net/http"
)

func (api *API) EnqueueTask(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("X-GitHub-Event") != "check_suite" {
		render.NoContent(w, r)
		return
	}

	data := github.CheckSuiteEvent{}
	if err := render.DecodeJSON(r.Body, &data); err != nil {
		E.SendError(w, r, nil, http.StatusBadRequest, "invalid input")
		return
	}

	var userID uint64
	err := api.DB.QueryRow(r.Context(), `
		SELECT "id" FROM "users" WHERE "github_id"=$1 AND "repository_id"=$2
	`, data.Sender.ID, data.Repo.ID).Scan(&userID)

	switch {
	case err == pgx.ErrNoRows:
		e := fmt.Errorf("user not found: %s (id=%d)", data.Sender.Login, data.Sender.ID)
		E.SendError(w, r, e, http.StatusBadRequest, e.Error())
		return
	case err != nil:
		E.Handle(w, r, err)
		return
	}

	appToken, err := api.App.Token()
	if err != nil {
		E.Handle(w, r, errors.Wrap(err, "could not get app token"))
		return
	}
	gh := github.New(appToken)
	err = gh.AuthAsInstallation(r.Context(), data.Inst.ID)
	if err != nil {
		E.Handle(w, r, errors.Wrap(err, "could not auth as installation"))
		return
	}

	checkRun, err := gh.CreateCheckRun(r.Context(), data.Repo.Owner.Login, data.Repo.Name, data.CheckSuite.Head)
	if err != nil {
		E.Handle(w, r, errors.Wrap(err, "could not create a check run"))
		return
	}

	err = db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {

		var commitID uint64

		err = tx.QueryRow(r.Context(), `
		INSERT INTO commits ("user_id", "commit", "check_run_id")
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, commit) DO NOTHING
		RETURNING "id"
		`, userID, data.CheckSuite.Head, checkRun.ID).Scan(&commitID)

		switch {
		case err == pgx.ErrNoRows: // conflict, the same commit
			return E.New(nil, http.StatusNoContent, "existing commit")
		case err != nil:
			return errors.WithStack(err)
		}

		_, err = tx.Exec(r.Context(), `INSERT INTO "tasks" ("commit_id") VALUES ($1);`, commitID)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	})
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	render.NoContent(w, r)
}

func (api *API) DequeueTask(w http.ResponseWriter, r *http.Request) {

	var (
		taskID                                  string
		commitID                                uint64
		instID                                  int
		commitHash, login, repoName, archiveURL string
	)

	err := db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {

		err := tx.QueryRow(r.Context(), `
		UPDATE tasks SET status='executing', started_at=STATEMENT_TIMESTAMP()
		WHERE id=(
			SELECT id FROM tasks
			WHERE status='enqueued'
			ORDER BY id
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		) RETURNING id, commit_id
		;`).Scan(&taskID, &commitID)

		switch {
		case err == pgx.ErrNoRows:
			return E.New(nil, http.StatusNoContent, "no enqueued tasks")
		case err != nil:
			return errors.WithStack(err)
		}

		err = api.DB.QueryRow(r.Context(), `
		SELECT u.login, c.commit, u.repository_name, u.installation_id
		FROM commits AS c JOIN users as u ON(u.id=c.user_id)
		WHERE c.id=$1
		;`, commitID).Scan(&login, &commitHash, &repoName, &instID)

		switch {
		case err == pgx.ErrNoRows:
			e := fmt.Errorf("enqueued commit not found: %d", commitID)
			return E.New(e, http.StatusNotFound, e.Error())
		case err != nil:
			return errors.WithStack(err)
		}

		appToken, err := api.App.Token()
		if err != nil {
			return errors.Wrap(err, "could not get app token")
		}
		gh := github.New(appToken)
		err = gh.AuthAsInstallation(r.Context(), instID)
		if err != nil {
			return errors.Wrap(err, "could not auth as installation")
		}

		archive, err := gh.Archive(r.Context(), login, repoName, commitHash)
		if err != nil {
			return errors.Wrap(err, "could not download archive")
		}

		s3Client := s3.New(api.AWS.Session(), api.AWS.Bucket)
		archiveKey := fmt.Sprintf("%s/%s/%s.zip", login, repoName, commitHash)
		err = s3Client.Upload(r.Context(), archiveKey, bytes.NewBuffer(archive))
		if err != nil {
			return errors.Wrap(err, "could not upload archive to S3")
		}

		archiveURL, err = s3Client.URL(r.Context(), archiveKey)
		if err != nil {
			return errors.Wrap(err, "could not get S3 URL")
		}
		return nil
	})

	if err != nil {
		E.Handle(w, r, err)
		return
	}

	render.JSON(w, r, &models.Task{
		Id:  taskID,
		Ref: fmt.Sprintf("%s:%s", login, commitHash[:8]),
		Url: archiveURL,
	})
}

func (api *API) FinishTask(w http.ResponseWriter, r *http.Request) {

	taskID := chi.URLParam(r, "taskID")
	if _, err := uuid.Parse(taskID); err != nil {
		E.SendError(w, r, err, http.StatusBadRequest, "invalid uuid")
		return
	}

	var (
		commitID  uint64
		isChecked bool
	)

	err := api.DB.QueryRow(r.Context(), `
	SELECT c.id, c.is_checked
	FROM commits AS c JOIN tasks AS t ON(c.id=t.commit_id)
	WHERE t.id=$1
	;`, taskID).Scan(&commitID, &isChecked)

	switch {
	case err == pgx.ErrNoRows:
		e := fmt.Errorf("unknown task: %v", taskID)
		E.SendError(w, r, e, http.StatusNotFound, e.Error())
		return
	case err != nil:
		E.Handle(w, r, err)
		return
	case isChecked:
		render.NoContent(w, r)
		return
	}

	var stages []models.Stage
	if err = render.DecodeJSON(r.Body, &stages); err != nil {
		E.SendError(w, r, err, http.StatusBadRequest, "invalid input")
		return
	}

	testNames := utils.UniqueStrings(stages, "Test")
	testIds, err := api.getTestIds(r.Context(), testNames)
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	var crows [][]interface{}
	for _, stage := range stages {
		var testID *uint64
		if stage.Test != "" {
			if v, ok := testIds[stage.Test]; ok {
				testID = &v
			} else {
				continue
			}
		}
		crows = append(crows, []interface{}{commitID, testID, stage.Name, stage.Status, stage.Output})
	}

	err = db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {

		cols := []string{"commit_id", "test_id", "name", "status", "output"}
		cfr := pgx.CopyFromRows(crows)
		_, err = tx.CopyFrom(r.Context(), pgx.Identifier{"checks"}, cols, cfr)
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = tx.Exec(r.Context(), `UPDATE commits SET is_checked='t' WHERE id=$1`, commitID)
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = tx.Exec(r.Context(), `
		UPDATE tasks SET status='finished', finished_at=STATEMENT_TIMESTAMP() WHERE id=$1
		`, taskID)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	})

	if err != nil {
		E.Handle(w, r, err)
		return
	}

	render.NoContent(w, r)
}
