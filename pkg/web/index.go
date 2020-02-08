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
	user, err := web.getUser(r)
	if err != nil {
		web.HandleError(w, err)
		return
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

func (web *Web) getUser(r *http.Request) (*models.User, error) {
	session, err := r.Cookie("session")
	if err != nil {
		return nil, nil
	}
	return web.API.GetUser(r.Context(), session.Value)
}
