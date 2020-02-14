package web

import (
	"net/http"
)

func (web *Web) GetQuickstart(w http.ResponseWriter, r *http.Request) {
	tpl, err := web.Templates.New("quickstart")
	if err != nil {
		web.HandleError(w, r, err)
		return
	}
	if err := web.Render(w, tpl, nil); err != nil {
		web.HandleError(w, r, err)
		return
	}
}
