package errors

import (
	"log"
	"net/http"

	"github.com/go-chi/render"
)

type APIError struct {
	Err  error
	Code int
	Msg  string
}

func (e *APIError) Error() string {
	return e.Msg
}

func (e *APIError) JSON() render.M {
	return render.M{
		"error":   http.StatusText(e.Code),
		"message": e.Msg,
	}
}

func New(err error, code int, msg string) *APIError {
	return &APIError{err, code, msg}
}

func Handle(w http.ResponseWriter, r *http.Request, err error) {
	switch v := err.(type) {
	case *APIError:
		render.Status(r, v.Code)
		render.JSON(w, r, v.JSON())
	default:
		log.Printf("[ERR] %v", err)
		e := New(err, http.StatusInternalServerError, "system error")
		render.Status(r, e.Code)
		render.JSON(w, r, e.JSON())
	}
}

func SendError(w http.ResponseWriter, r *http.Request, err error, code int, msg string) {
	Handle(w, r, New(err, code, msg))
}
