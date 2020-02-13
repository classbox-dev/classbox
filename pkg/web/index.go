package web

import (
	"github.com/mkuznets/classbox/pkg/api/models"
	"net/http"
)

type indexPage struct {
	User    *models.User
	DocsURL string
	Tests   []*models.Test
	Stats   []*models.UserStat
}

func (web *Web) GetIndex(w http.ResponseWriter, r *http.Request) {

	tests, err := web.API.GetTests(r.Context())
	if err != nil {
		web.HandleError(w, err)
		return
	}
	stats, err := web.API.GetUserStats(r.Context())
	if err != nil {
		web.HandleError(w, err)
		return
	}

	var user *models.User
	if v, ok := r.Context().Value("User").(*models.User); ok {
		user = v
	}

	tpl, err := web.Templates.New("index")
	if err != nil {
		web.HandleError(w, err)
		return
	}

	page := &indexPage{
		User:    user,
		DocsURL: web.DocsURL,
		Tests:   tests,
		Stats:   stats,
	}

	if err := web.Render(w, tpl, page); err != nil {
		web.HandleError(w, err)
		return
	}
}
