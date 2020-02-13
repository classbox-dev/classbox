package web

import (
	"net/http"
	"strconv"
)

func (web *Web) GetSignin(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Query().Get("step") {

	default:
		url, err := web.API.GetAppUrl(r.Context())
		if err != nil {
			web.handleSigninError(w, r, err)
			return
		}
		http.Redirect(w, r, url, http.StatusFound)
		return

	case "signin":
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		stage, err := web.API.Signin(r.Context(), code, state)
		if err != nil {
			web.handleSigninError(w, r, err)
			return
		}
		stage.SetAuthCookie(w)
		http.Redirect(w, r, stage.Url, http.StatusFound)
		return

	case "create":
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		stage, err := web.API.CreateUser(r.Context(), code, state)
		if err != nil {
			web.handleSigninError(w, r, err)
			return
		}
		stage.SetAuthCookie(w)
		http.Redirect(w, r, stage.Url, http.StatusFound)
		return

	case "install":
		instId, err := strconv.ParseUint(r.URL.Query().Get("installation_id"), 10, 64)
		if err != nil {
			web.handleSigninError(w, r, err)
			return
		}
		state := r.URL.Query().Get("state")

		stage, err := web.API.InstallApp(r.Context(), instId, state)
		if err != nil {
			web.handleSigninError(w, r, err)
			return
		}
		stage.SetAuthCookie(w)
		http.Redirect(w, r, stage.Url, http.StatusFound)
		return

	case "honor_code":
		tpl, err := web.Templates.New("honor_code")
		if err != nil {
			web.handleSigninError(w, r, err)
			return
		}
		if err := web.Render(w, tpl, nil); err != nil {
			web.handleSigninError(w, r, err)
			return
		}
	}
}

func (web *Web) handleSigninError(w http.ResponseWriter, r *http.Request, e error) {
	tpl, err := web.Templates.New("signin_error")
	if err != nil {
		web.HandleError(w, r, err)
		return
	}
	if err := web.Render(w, tpl, e.Error()); err != nil {
		web.HandleError(w, r, err)
		return
	}
}

func (web *Web) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		MaxAge: -1,
	})
	http.Redirect(w, r, web.WebURL, http.StatusFound)
}
