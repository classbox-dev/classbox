package web

import (
	"net/http"
)

func (web *Web) GetIndex(w http.ResponseWriter, r *http.Request) {

	tests, err := web.API.GetTests(r.Context())
	if err != nil {
		web.HandleError(w, err)
		return
	}

	tpl, err := web.Templates.New("index")
	if err != nil {
		web.HandleError(w, err)
		return
	}

	if err := web.Render(w, tpl, tests); err != nil {
		web.HandleError(w, err)
		return
	}
}
