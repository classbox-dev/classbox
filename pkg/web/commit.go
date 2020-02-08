package web

import (
	"github.com/go-chi/chi"
	"net/http"
)

func (web *Web) GetCommit(w http.ResponseWriter, r *http.Request) {

	commitHash := chi.URLParam(r, "commitHash")
	login := chi.URLParam(r, "login")

	commit, err := web.API.GetCommit(r.Context(), login, commitHash)
	if err != nil {
		web.HandleError(w, err)
		return
	}

	tpl, err := web.Templates.New("commit")
	if err != nil {
		web.HandleError(w, err)
		return
	}

	if err := web.Render(w, tpl, commit); err != nil {
		web.HandleError(w, err)
		return
	}
}
