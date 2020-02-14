package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v4"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/mkuznets/classbox/pkg/api/models"
	"github.com/mkuznets/classbox/pkg/db"
	"github.com/mkuznets/classbox/pkg/github"
	"github.com/mkuznets/classbox/pkg/utils"
	"github.com/pkg/errors"
	"log"
	"net/http"
)

const repoName = "hsecode-stdlib"

func (api *API) AppURL(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"url": api.App.Config().AuthCodeURL(api.RandomState)})
}

func (api *API) OAuthURL(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"url": api.OAuth.Config().AuthCodeURL(api.RandomState)})
}

type oauthData struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func (api *API) Signin(w http.ResponseWriter, r *http.Request) {

	data := oauthData{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		E.SendError(w, r, err, http.StatusBadRequest, "invalid input")
		return
	}

	if api.RandomState != data.State {
		E.SendError(w, r, nil, http.StatusBadRequest, "invalid state")
		return
	}

	token, err := api.App.Config().Exchange(r.Context(), data.Code)
	if err != nil {
		E.Handle(w, r, errors.Wrap(err, "could not get token"))
		return
	}

	redirectToOAuth := func() {
		render.JSON(w, r, models.AuthStage{Url: api.OAuth.Config().AuthCodeURL(api.RandomState)})
	}

	gh := github.New(token)

	user, err := gh.User(r.Context())
	if err != nil {
		E.Handle(w, r, errors.Wrap(err, "user request error"))
		return
	}

	repo, err := gh.Repo(r.Context(), user.Login, repoName)
	if err != nil {
		if e, ok := err.(*github.ErrorResponse); ok && e.NotFound() {
			redirectToOAuth()
			return
		} else {
			E.Handle(w, r, errors.Wrap(err, "repo request error"))
			return
		}
	}

	err = db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {
		var (
			userId         uint64
			installationId *uint64
		)
		err := tx.QueryRow(r.Context(), `
		UPDATE "users" SET login=$2, email=$3, repository_id=$4, repository_name=$5
		WHERE "github_id"=$1
		RETURNING id, installation_id
		`, user.ID, user.Login, user.Email, repo.ID, repo.Name).Scan(&userId, &installationId)
		switch {
		case err == pgx.ErrNoRows:
			redirectToOAuth()
			return nil
		case err != nil:
			return errors.WithStack(err)
		}

		found := false
		if installationId != nil {
			repos, err := gh.ReposByInstID(r.Context(), *installationId)
			if err != nil {
				return errors.WithStack(err)
			}
			for _, r := range repos {
				if r.ID == repo.ID {
					found = true
				}
			}
		}
		if !found {
			redirectToOAuth()
			return nil
		}
		session, err := createSession(r.Context(), tx, userId)
		if err != nil {
			return errors.WithStack(err)
		}
		render.JSON(w, r, &models.AuthStage{Session: session, Url: api.WebUrl})
		return nil
	})
	if err != nil {
		E.Handle(w, r, err)
		return
	}
}

func (api *API) CreateUser(w http.ResponseWriter, r *http.Request) {

	data := oauthData{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		E.SendError(w, r, err, http.StatusBadRequest, "invalid input")
		return
	}

	if api.RandomState != data.State {
		E.SendError(w, r, nil, http.StatusBadRequest, "invalid state")
		return
	}

	token, err := api.OAuth.Config().Exchange(r.Context(), data.Code)
	if err != nil {
		E.Handle(w, r, errors.Wrap(err, "could not get token"))
		return
	}

	gh := github.New(token)

	user, err := gh.User(r.Context())
	if err != nil {
		E.Handle(w, r, errors.Wrap(err, "user request error"))
		return
	}

	repo, err := gh.Repo(r.Context(), user.Login, repoName)
	if err != nil {
		if e, ok := err.(*github.ErrorResponse); ok && e.NotFound() {
			repo, err = gh.CreateRepoFromTemplate(r.Context(), "mkuznets/stdlib-template", repoName, true)
			if err != nil {
				E.Handle(w, r, errors.Wrap(err, "could not create a repo"))
				return
			}
		} else {
			E.Handle(w, r, errors.Wrap(err, "repo request error"))
			return
		}
	}

	var (
		userId    uint64
		instId    *uint64
		honorCode bool
	)
	err = db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {
		err := tx.QueryRow(r.Context(), `
		INSERT INTO users ("github_id", "login", "email", "repository_id", "repository_name")
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT ("github_id") DO UPDATE
		SET
			email=EXCLUDED.email,
			repository_id=EXCLUDED.repository_id,
			repository_name=EXCLUDED.repository_name,
			login=EXCLUDED.login
		RETURNING id, honor_code, installation_id
		`, user.ID, user.Login, user.Email, repo.ID, repo.Name).Scan(&userId, &honorCode, &instId)
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	})
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	err = gh.RevokeOAuth(r.Context(), api.OAuth.ClientID, api.OAuth.ClientSecret)
	if err != nil {
		log.Printf("[WARN] could not revoke oauth authorization: %v", err)
	}

	appToken, err := api.App.Token()
	if err != nil {
		E.Handle(w, r, errors.Wrap(err, "could not get app token"))
		return
	}

	redirectToFinish := func() {
		var session string
		err = db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {
			session, err = createSession(r.Context(), tx, userId)
			if err != nil {
				return errors.WithStack(err)
			}
			return nil
		})
		if err != nil {
			E.Handle(w, r, errors.Wrap(err, "could not create session"))
			return
		}

		var finalPath string
		if honorCode {
			finalPath = "/"
		} else {
			finalPath = "/signin?step=honor_code"
		}
		finishUrl := fmt.Sprintf("%s%s", api.WebUrl, finalPath)
		render.JSON(w, r, models.AuthStage{
			Session: session,
			Url:     finishUrl,
		})
	}

	redirectToInstall := func() {
		installUrl := fmt.Sprintf("https://github.com/apps/%s/installations/new/permissions"+
			"?suggested_target_id=%d&repository_ids[]=%d&state=%s", api.App.Name, user.ID, repo.ID, api.RandomState)
		render.JSON(w, r, models.AuthStage{
			Url: installUrl,
		})
	}

	app := github.New(appToken)

	inst, err := app.InstallationByLogin(r.Context(), user.Login)
	if err != nil {
		if e, ok := err.(*github.ErrorResponse); ok && e.NotFound() {
			redirectToInstall()
			return
		}
		E.Handle(w, r, errors.Wrap(err, "could not check installation"))
		return
	}

	err = db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {
		_, err = tx.Exec(r.Context(), `
		UPDATE "users" SET installation_id=$1 WHERE "id"=$2`, inst.ID, userId)
		if err != nil {
			return errors.WithStack(err)
		}
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	})

	if err := app.AuthAsInstallation(r.Context(), inst.ID); err != nil {
		E.Handle(w, r, errors.Wrap(err, "could not auth as installation"))
		return
	}

	repos, err := app.InstallationRepos(r.Context())
	if err != nil {
		E.Handle(w, r, errors.Wrap(err, "could not list installation repos"))
		return
	}

	for _, r := range repos {
		if r.ID == repo.ID {
			redirectToFinish()
			return
		}
	}

	app = github.New(appToken)
	if err := app.Uninstall(r.Context(), inst.ID); err != nil {
		log.Printf("[WARN] could not uninstall the app: %v", err)
	}
	redirectToInstall()
}

func (api *API) InstallApp(w http.ResponseWriter, r *http.Request) {

	data := models.AppInstallData{}
	if err := render.DecodeJSON(r.Body, &data); err != nil {
		E.SendError(w, r, nil, http.StatusBadRequest, "invalid input")
		return
	}

	if api.RandomState != data.State {
		E.SendError(w, r, nil, http.StatusBadRequest, "invalid state")
		return
	}

	appToken, err := api.App.Token()
	if err != nil {
		E.Handle(w, r, errors.Wrap(err, "could not get app token"))
		return
	}
	app := github.New(appToken)

	inst, err := app.InstallationByID(r.Context(), data.InstID)
	if err != nil {
		E.SendError(w, r, err, http.StatusNotFound, "installation not found")
		return
	}

	var (
		login, repoName string
		userId          uint64
		honorCode       bool
	)
	err = api.DB.QueryRow(r.Context(), `
	SELECT id, login, repository_name, honor_code FROM "users" WHERE "github_id"=$1 LIMIT 1
	`, inst.Account.ID).Scan(&userId, &login, &repoName, &honorCode)
	switch {
	case err == pgx.ErrNoRows:
		e := fmt.Errorf("user not found: %s (id=%d)", inst.Account.Login, inst.Account.ID)
		E.SendError(w, r, e, http.StatusBadRequest, e.Error())
		return
	case err != nil:
		E.Handle(w, r, err)
		return
	}

	var session string
	err = db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {
		_, err = tx.Exec(r.Context(), `
		UPDATE "users" SET installation_id=$1, honor_code='t' WHERE "id"=$2
		`, inst.ID, userId)
		if err != nil {
			return errors.WithStack(err)
		}
		session, err = createSession(r.Context(), tx, userId)
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	})

	if err != nil {
		E.Handle(w, r, err)
		return
	}

	var finalPath string
	if honorCode {
		finalPath = "/"
	} else {
		finalPath = "/signin?step=honor_code"
	}
	finishUrl := fmt.Sprintf("%s%s", api.WebUrl, finalPath)
	render.JSON(w, r, &models.AuthStage{Session: session, Url: finishUrl})
}

func createSession(ctx context.Context, tx pgx.Tx, userId uint64) (string, error) {
	session := utils.RandomString(96)
	_, err := tx.Exec(ctx, `
		INSERT INTO sessions (user_id, session, expires_at)
		VALUES ($1, $2, STATEMENT_TIMESTAMP() + interval '30 days')
		`, userId, session)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return session, nil
}
