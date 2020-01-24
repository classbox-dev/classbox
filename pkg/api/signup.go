package api

import (
	"bytes"
	"fmt"
	"github.com/go-chi/render"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/pkg/errors"
	"net/http"
)

// Scoreboard returns scores of all students
func (api *API) Signup(w http.ResponseWriter, r *http.Request) {

	code := r.FormValue("code")
	if code == "" {
		url := api.OAuth.AuthCodeURL(api.RandomState)
		http.Redirect(w, r, url, http.StatusFound)
		return
	} else {
		if api.RandomState != r.FormValue("state") {
			E.Render(w, r, E.Internal(fmt.Errorf("invalid state: %v vs %v", api.RandomState, r.FormValue("state"))))
			return
		}

		token, err := api.OAuth.Exchange(r.Context(), r.FormValue("code"))
		if err != nil {
			E.Render(w, r, E.Internal(errors.Wrap(err, "could not get token")))
			return
		}

		url := fmt.Sprintf("https://api.github.com/applications/%s/grant", api.OAuth.ClientID)
		jsonStr := []byte(fmt.Sprintf(`{"access_token":"%s"}`, token.AccessToken))
		req, err := http.NewRequestWithContext(r.Context(), "DELETE", url, bytes.NewBuffer(jsonStr))
		if err != nil {
			E.Render(w, r, E.Internal(errors.Wrap(err, "request error")))
			return
		}
		req.Header.Set("Accept", "application/vnd.github.doctor-strange-preview+json")
		req.SetBasicAuth(api.OAuth.ClientID, api.OAuth.ClientSecret)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			E.Render(w, r, E.Internal(errors.Wrap(err, "request error")))
			return
		}
		//noinspection GoUnhandledErrorResult
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			E.Render(w, r, E.Internal(fmt.Errorf("could not delete authorisation")))
			return
		}

		render.JSON(w, r, map[string]string{"status": "ok"})
	}
}
