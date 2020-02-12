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
	"github.com/mkuznets/classbox/pkg/web"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"time"
)

func (api *API) EnqueueTask(w http.ResponseWriter, r *http.Request) {
	eventName := r.Header.Get("X-GitHub-Event")

	if eventName != "check_suite" {
		render.NoContent(w, r)
		return
	}

	cs := github.CheckSuiteEvent{}
	if err := render.DecodeJSON(r.Body, &cs); err != nil {
		E.SendError(w, r, nil, http.StatusBadRequest, "invalid input")
		return
	}
	if cs.Action != "requested" && cs.Action != "rerequested" {
		render.NoContent(w, r)
		return
	}

	var userID uint64
	err := api.DB.QueryRow(r.Context(), `
	SELECT "id" FROM "users" WHERE "github_id"=$1 AND "repository_id"=$2 LIMIT 1
	`, cs.Sender.ID, cs.Repo.ID).Scan(&userID)

	switch {
	case err == pgx.ErrNoRows:
		e := fmt.Errorf("user not found: %s (id=%d)", cs.Sender.Login, cs.Sender.ID)
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
	if err := gh.AuthAsInstallation(r.Context(), cs.Inst.ID); err != nil {
		E.Handle(w, r, errors.Wrap(err, "could not auth as installation"))
		return
	}

	checkRun, err := gh.CreateCheckRun(
		r.Context(), cs.Repo.Owner.Login, cs.Repo.Name,
		&github.CheckRun{
			Name:   "stdlib tests",
			Commit: cs.CheckSuite.Head,
			Status: "queued",
			Url:    fmt.Sprintf("%s/commit/%s:%s", api.WebUrl, cs.Repo.Owner.Login, cs.CheckSuite.Head),
		},
	)
	if err != nil {
		E.Handle(w, r, errors.Wrap(err, "could not create a check run"))
		return
	}

	err = db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {

		var commitID uint64

		err = tx.QueryRow(r.Context(), `
		INSERT INTO commits ("user_id", "commit", "check_run_id")
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, commit) DO UPDATE
		SET check_run_id=EXCLUDED.check_run_id, is_checked='f'
		RETURNING "id"
		`, userID, cs.CheckSuite.Head, checkRun.ID).Scan(&commitID)

		switch {
		case err == pgx.ErrNoRows: // conflict, the same commit
			return E.New(nil, http.StatusNoContent, "existing commit")
		case err != nil:
			return errors.WithStack(err)
		}

		_, err = tx.Exec(r.Context(), `
		INSERT INTO "tasks" ("commit_id") VALUES ($1) ON CONFLICT (commit_id) DO UPDATE SET status='enqueued';`, commitID)
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = tx.Exec(r.Context(), `
		DELETE FROM "checks" WHERE commit_id=$1;`, commitID)
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
		checkRunId                              uint64
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
		SELECT u.login, c.commit, u.repository_name, u.installation_id, c.check_run_id
		FROM commits AS c JOIN users as u ON(u.id=c.user_id)
		WHERE c.id=$1 LIMIT 1
		;`, commitID).Scan(&login, &commitHash, &repoName, &instID, &checkRunId)

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
		if err := gh.AuthAsInstallation(r.Context(), instID); err != nil {
			return errors.Wrap(err, "could not auth as installation")
		}

		checkRun := &github.CheckRun{
			ID:        checkRunId,
			Status:    "in_progress",
			StartTime: time.Now().UTC().Format(time.RFC3339),
		}
		if err := gh.UpdateCheckRun(r.Context(), login, repoName, checkRun); err != nil {
			return errors.Wrap(err, "could not update check run")
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
		commitId    uint64
		isChecked   bool
		checkRun    github.CheckRun
		login, repo string
		instId      int
	)

	err := api.DB.QueryRow(r.Context(), `
	SELECT c.id, c.is_checked, c.check_run_id, u.login, u.repository_name, u.installation_id
	FROM
		commits AS c
		JOIN tasks AS t ON(c.id=t.commit_id)
		JOIN users AS u ON (u.id=c.user_id)
	WHERE t.id=$1 LIMIT 1
	;`, taskID).Scan(&commitId, &isChecked, &checkRun.ID, &login, &repo, &instId)
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

	var stages []*models.Stage
	if err = render.DecodeJSON(r.Body, &stages); err != nil {
		E.SendError(w, r, err, http.StatusBadRequest, "invalid input")
		return
	}
	if len(stages) == 0 {
		E.SendError(w, r, err, http.StatusBadRequest, "stage list cannot be empty")
	}

	testNames := utils.UniqueStringFields(stages, "Test")
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
		crows = append(crows, []interface{}{commitId, testID, stage.Name, stage.Status, stage.Output})
	}

	title := ""
	failures := []string{}
	for _, s := range stages {
		if s.Status != "success" {
			failures = append(failures, s.Name)
		}
	}
	checkRun.Status = "completed"
	checkRun.CompletionTime = time.Now().UTC().Format(time.RFC3339)
	if len(failures) > 0 {
		checkRun.Conclusion = "failure"
		title = fmt.Sprintf("Failed: %s", strings.Join(failures, " , "))
	} else {
		checkRun.Conclusion = "success"
		title = "Success"
	}

	ts, err := web.NewTemplates()
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	tpl, err := ts.New("check_run")
	if err != nil {
		E.Handle(w, r, err)
		return
	}
	summary := bytes.NewBufferString("")
	if err := tpl.ExecuteTemplate(summary, "markdown", stages); err != nil {
		E.Handle(w, r, err)
		return
	}
	checkRun.Output = &github.CheckRunOutput{
		Title:   title,
		Summary: summary.String(),
	}

	err = db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {

		cols := []string{"commit_id", "test_id", "name", "status", "output"}
		cfr := pgx.CopyFromRows(crows)
		_, err = tx.CopyFrom(r.Context(), pgx.Identifier{"checks"}, cols, cfr)
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = tx.Exec(r.Context(), `UPDATE commits SET is_checked='t' WHERE id=$1`, commitId)
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = tx.Exec(r.Context(), `
		UPDATE tasks SET status='finished', finished_at=STATEMENT_TIMESTAMP() WHERE id=$1
		`, taskID)
		if err != nil {
			return errors.WithStack(err)
		}

		appToken, err := api.App.Token()
		if err != nil {
			return errors.Wrap(err, "could not get app token")
		}
		gh := github.New(appToken)
		if err := gh.AuthAsInstallation(r.Context(), instId); err != nil {
			return errors.Wrap(err, "could not authenticate as installation")
		}
		if err := gh.UpdateCheckRun(r.Context(), login, repo, &checkRun); err != nil {
			return errors.Wrap(err, "could not finalise check run")
		}

		return nil
	})
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	render.NoContent(w, r)
}
