package web

import (
	"net/http"
)

type quickstartPage struct {
	DocsURL string
}

func (web *Web) GetQuickstart(w http.ResponseWriter, r *http.Request) {
	tpl, err := web.Templates.New("quickstart")
	if err != nil {
		web.HandleError(w, r, err)
		return
	}
	page := quickstartPage{DocsURL: web.DocsURL}
	if err := web.Render(w, tpl, &page); err != nil {
		web.HandleError(w, r, err)
		return
	}
}
