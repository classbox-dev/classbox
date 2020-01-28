package api

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/go-chi/render"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/mkuznets/classbox/pkg/github"
	"github.com/mkuznets/classbox/pkg/s3"
	"github.com/pkg/errors"
	"net/http"
)

type taskResponse struct {
	Id      string `json:"id"`
	Login   string `json:"login"`
	Commit  string `json:"commit"`
	Archive string `json:"archive"`
}

func (api *API) Pop(w http.ResponseWriter, r *http.Request) {

	tx, err := api.DB.Begin(r.Context())
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "could not start transaction")))
		return
	}
	//noinspection GoUnhandledErrorResult
	defer tx.Rollback(r.Context())

	var (
		taskID   string
		commitID uint64
	)

	err = tx.QueryRow(r.Context(), `
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
	case err == sql.ErrNoRows:
		w.WriteHeader(http.StatusNoContent)
		return
	case err != nil:
		E.Render(w, r, E.Internal(errors.Wrap(err, "query error")))
		return
	}

	var (
		instID                      int
		commitHash, login, repoName string
	)

	err = api.DB.QueryRow(r.Context(), `
	SELECT u.login, c.commit, u.repository_name, u.installation_id
	FROM commits AS c JOIN users as u ON(u.id=c.user_id)
    WHERE c.id=$1
	;`, commitID).Scan(&login, &commitHash, &repoName, &instID)

	switch {
	case err == sql.ErrNoRows:
		E.Render(w, r, E.Internal(fmt.Errorf("enqueued commit not found: %d", commitID)))
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
	err = gh.AuthAsInstallation(r.Context(), instID)
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "could not auth as installation")))
		return
	}

	archive, err := gh.Archive(r.Context(), login, repoName, commitHash)
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "could not download archive")))
		return
	}

	s3Client := s3.New(api.AWS.Session(), api.AWS.Bucket)
	archiveKey := fmt.Sprintf("%s/%s/%s.zip", login, repoName, commitHash)
	err = s3Client.Upload(r.Context(), archiveKey, bytes.NewBuffer(archive))
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "could not upload archive to S3")))
		return
	}

	archiveURL, err := s3Client.URL(r.Context(), archiveKey)
	if err != nil {
		E.Render(w, r, E.Internal(errors.Wrap(err, "could not get S3 URL")))
		return
	}

	resp := taskResponse{
		Id:      taskID,
		Login:   login,
		Commit:  commitHash,
		Archive: archiveURL,
	}
	render.JSON(w, r, &resp)
}
