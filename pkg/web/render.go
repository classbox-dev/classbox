package web

import (
	"bytes"
	"html/template"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

func (web *Web) Render(w http.ResponseWriter, tpl *template.Template, v interface{}) error {
	resp := bytes.NewBufferString("")
	if err := tpl.Execute(resp, v); err != nil {
		return errors.WithStack(err)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := io.Copy(w, resp); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
