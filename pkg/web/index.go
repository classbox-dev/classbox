package web

import (
	"math"
	"net/http"

	"github.com/mkuznets/classbox/pkg/api/models"
)

type indexPage struct {
	User    *models.User
	DocsURL string
	Stats   *models.UserStats
	Grade   float64
}

func gradeFromScore(score uint64) float64 {
	preGrade := float64(score) / 10.
	grade := math.Min(10., preGrade)
	return grade
}

func (web *Web) GetIndex(w http.ResponseWriter, r *http.Request) {

	var user *models.User
	if v, ok := r.Context().Value("User").(*models.User); ok {
		user = v
	}

	page := &indexPage{
		User:    user,
		DocsURL: web.DocsURL,
	}

	var tplName string
	switch user {
	case nil:
		tplName = "index_landing"
	default:
		tplName = "index_user"

		stats, err := web.API(r).GetUserStats(r.Context())
		if err != nil {
			web.HandleError(w, r, err)
			return
		}
		page.Stats = stats
		page.Grade = gradeFromScore(stats.Score)
	}

	tpl, err := web.Templates.New(tplName)
	if err != nil {
		web.HandleError(w, r, err)
		return
	}
	if err := web.Render(w, tpl, page); err != nil {
		web.HandleError(w, r, err)
		return
	}
}
