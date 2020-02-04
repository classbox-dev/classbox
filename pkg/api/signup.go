package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v4"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/mkuznets/classbox/pkg/db"
	"github.com/mkuznets/classbox/pkg/github"
	"github.com/pkg/errors"
	"log"
	"net/http"
)

const repoName = "hsecode-stdlib"

func (api *API) OAuthURL(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{"url": api.OAuth.AuthCodeURL(api.RandomState)})
}

type oauthData struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

func (api *API) CreateUser(w http.ResponseWriter, r *http.Request) {

	data := oauthData{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		E.Render(w, r, E.BadRequest(errors.Wrap(err, "invalid input")))
		return
	}

	if api.RandomState != data.State {
		E.Render(w, r, E.BadRequest(fmt.Errorf("invalid state")))
		return
	}

	token, err := api.OAuth.Exchange(r.Context(), data.Code)
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

	err = db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {
		_, err = tx.Exec(r.Context(), `
		INSERT INTO users ("github_id", "login", "email", "repository_id", "repository_name")
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT ("github_id") DO UPDATE
		SET
			email=EXCLUDED.email,
			repository_id=EXCLUDED.repository_id,
			repository_name=EXCLUDED.repository_name,
			login=EXCLUDED.login
		`, user.ID, user.Login, user.Email, repo.ID, repo.Name)
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	})
	if err != nil {
		E.Handle(w, r, err)
		return
	}

	// err = gh.RevokeOAuth(r.Context(), api.OAuth.ClientID, api.OAuth.ClientSecret)
	// if err != nil {
	// 	log.Printf("[WARN] could not revoke oauth authorization: %v", err)
	// }

	appToken, err := api.App.Token()
	if err != nil {
		E.Handle(w, r, errors.Wrap(err, "could not get app token"))
		return
	}
	app := github.New(appToken)
	inst, err := app.InstallationByLogin(r.Context(), user.Login)
	if err == nil {
		err := app.Uninstall(r.Context(), inst.ID)
		if err != nil {
			log.Printf("[WARN] could not uninstall the app: %v", err)
		}
	}

	url := fmt.Sprintf("https://github.com/apps/hsecode/installations/new/permissions"+
		"?suggested_target_id=%d&repository_ids[]=%d&state=%s", user.ID, repo.ID, api.RandomState)

	render.JSON(w, r, render.M{"url": url})
}

type appData struct {
	InstID int    `json:"installation_id"`
	State  string `json:"state"`
}

func (api *API) InstallApp(w http.ResponseWriter, r *http.Request) {

	data := appData{}
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
		E.Render(w, r, E.NotFound(errors.Wrap(err, "installation not found")))
		return
	}

	var login, repoName string
	err = api.DB.QueryRow(r.Context(), `SELECT "login", "repository_name" FROM "users" WHERE "github_id"=$1`, inst.Account.ID).Scan(&login, &repoName)
	switch {
	case err == pgx.ErrNoRows:
		e := fmt.Errorf("user not found: %s (id=%d)", inst.Account.Login, inst.Account.ID)
		E.SendError(w, r, e, http.StatusBadRequest, e.Error())
		return
	case err != nil:
		E.Handle(w, r, err)
		return
	}

	err = db.Tx(r.Context(), api.DB, func(tx pgx.Tx) error {
		_, err = tx.Exec(r.Context(), `
		UPDATE "users" SET "installation_id"=$1 WHERE "github_id"=$2
		`, inst.ID, inst.Account.ID)
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	})

	if err != nil {
		E.Handle(w, r, err)
		return
	}

	render.JSON(w, r, render.M{"repo": fmt.Sprintf("%s/%s", login, repoName)})
}
