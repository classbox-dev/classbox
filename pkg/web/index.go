package web

import (
	"github.com/mkuznets/classbox/pkg/api/models"
	"net/http"
)

type page struct {
	Tests []*models.Test
	Stats []*models.UserStat
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

	tpl, err := web.Templates.New("index")
	if err != nil {
		web.HandleError(w, err)
		return
	}

	if err := web.Render(w, tpl, page{tests, stats}); err != nil {
		web.HandleError(w, err)
		return
	}
}
