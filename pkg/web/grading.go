package web

import (
	"github.com/mkuznets/classbox/pkg/api/models"
	"net/http"
)

type gradingPage struct {
	User *models.User
}

func (web *Web) GetGrading(w http.ResponseWriter, r *http.Request) {
	tpl, err := web.Templates.New("grading")
	if err != nil {
		web.HandleError(w, r, err)
		return
	}
	if err := web.Templates.EnableMath(tpl); err != nil {
		web.HandleError(w, r, err)
		return
	}

	var user *models.User
	if v, ok := r.Context().Value("User").(*models.User); ok {
		user = v
	}
	page := gradingPage{User: user}

	if err := web.Render(w, tpl, &page); err != nil {
		web.HandleError(w, r, err)
		return
	}
}
