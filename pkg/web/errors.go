package web

import (
	"github.com/getsentry/sentry-go"
	"github.com/mkuznets/classbox/pkg/api/client"
	"github.com/mkuznets/classbox/pkg/api/models"
	"log"
	"net/http"
)

const systemErrorText = `Unexpected system error. Developers have been alerted and will handle the issue as soon as possible.`

func (web *Web) HandleError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	switch v := err.(type) {
	case client.ErrorResponse:
		web.SendError(w, r, v.Code, v.Message)
	default:
		if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				if user, ok := r.Context().Value("User").(*models.User); ok {
					scope.SetUser(sentry.User{Username: user.Login})
				}
				hub.CaptureException(err)
			})
		}
		log.Printf("[ERR] %v", v)
		web.SendError(w, r, http.StatusInternalServerError, systemErrorText)
	}
}

func (web *Web) SendError(w http.ResponseWriter, r *http.Request, code int, text string) {
	http500 := func(e error) {
		if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
			hub.CaptureException(e)
		}
		log.Printf("[ERR] could not send error: %v", e)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		//noinspection GoUnhandledErrorResult
		w.Write([]byte("internal system error")) // nolint
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

func (web *Web) NotFound(w http.ResponseWriter, r *http.Request) {
	web.SendError(w, r, http.StatusNotFound, "Page not found")
}
