package web

import (
	"github.com/mkuznets/classbox/pkg/api/models"
	"net/http"
)

type scoreboardPage struct {
	User  *models.User
	Stats []*models.UserStat
}

func (web *Web) GetScoreboard(w http.ResponseWriter, r *http.Request) {
	stats, err := web.API.GetUserStats(r.Context())
	if err != nil {
		web.HandleError(w, r, err)
		return
	}
	var user *models.User
	if v, ok := r.Context().Value("User").(*models.User); ok {
		user = v
	}
	tpl, err := web.Templates.New("scoreboard")
	if err != nil {
		web.HandleError(w, r, err)
		return
	}
	if err := web.Render(w, tpl, &scoreboardPage{user, stats}); err != nil {
		web.HandleError(w, r, err)
		return
	}
}
