package web

import (
	"github.com/mkuznets/classbox/pkg/api/client"
	"log"
	"net/http"
)

func (web *Web) HandleError(w http.ResponseWriter, err error) {
	if err != nil {
		if v, ok := err.(client.ErrorResponse); ok {
			web.SendError(w, v.Code, v.Message)
			return
		} else {
			web.SendError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
}

func (web *Web) SendError(w http.ResponseWriter, code int, text string) {
	http500 := func(e error) {
		log.Printf("[ERR] could not send error: %v", e)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		//noinspection GoUnhandledErrorResult
		w.Write([]byte("internal system error"))
	}
	tpl, err := web.Templates.New("error")
	if err != nil {
		http500(err)
		return
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tpl.Execute(w, text)
	if err != nil {
		http500(err)
		return
	}
}
