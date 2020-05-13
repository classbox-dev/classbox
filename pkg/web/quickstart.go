package web

import (
	"net/http"

	"github.com/mkuznets/classbox/pkg/api/models"
)

type quickstartPage struct {
	DocsURL string
	User    *models.User
}

func (web *Web) GetQuickstart(w http.ResponseWriter, r *http.Request) {
	tpl, err := web.Templates.New("quickstart")
	if err != nil {
		web.HandleError(w, r, err)
		return
	}

	var user *models.User
	if v, ok := r.Context().Value("User").(*models.User); ok {
		user = v
	}
	page := quickstartPage{DocsURL: web.DocsURL, User: user}

	if err := web.Render(w, tpl, &page); err != nil {
		web.HandleError(w, r, err)
		return
	}
}
