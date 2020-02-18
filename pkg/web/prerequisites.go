package web

import (
	"github.com/mkuznets/classbox/pkg/api/models"
	"net/http"
)

type prerequisitesPage struct {
	User *models.User
}

func (web *Web) GetPrerequisites(w http.ResponseWriter, r *http.Request) {
	tpl, err := web.Templates.New("prerequisites")
	if err != nil {
		web.HandleError(w, r, err)
		return
	}

	var user *models.User
	if v, ok := r.Context().Value("User").(*models.User); ok {
		user = v
	}
	page := prerequisitesPage{User: user}

	if err := web.Render(w, tpl, page); err != nil {
		web.HandleError(w, r, err)
		return
	}
}
