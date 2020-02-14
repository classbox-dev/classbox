package web

import (
	"net/http"
)

func (web *Web) GetPrerequisites(w http.ResponseWriter, r *http.Request) {
	tpl, err := web.Templates.New("prerequisites")
	if err != nil {
		web.HandleError(w, r, err)
		return
	}
	if err := web.Render(w, tpl, nil); err != nil {
		web.HandleError(w, r, err)
		return
	}
}
