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

// ErrResponse renderer type for handling all sorts of errors.
type ErrResponse struct {
	Err        error  `json:"-"`       // low-level runtime error
	StatusCode int    `json:"-"`       // http response status code
	ErrorText  string `json:"error"`   // user-level status message
	Details    string `json:"details"` // user-level status message
}

// Render adds HTTP status to the response
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.StatusCode)
	return nil
}

// BadRequest is a
func BadRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:        err,
		StatusCode: http.StatusBadRequest,
		ErrorText:  "Invalid Data",
		Details:    err.Error(),
	}
}

func NotFound(err error) render.Renderer {
	return &ErrResponse{
		Err:        err,
		StatusCode: http.StatusNotFound,
		ErrorText:  "Not Found",
		Details:    err.Error(),
	}
}

// Internal responds with an internal server error
func Internal(err error) render.Renderer {
	return &ErrResponse{
		Err:        err,
		StatusCode: http.StatusInternalServerError,
		ErrorText:  "Internal Server Error",
		Details:    err.Error(),
	}
}

// Render renders an error
func Render(w http.ResponseWriter, r *http.Request, v render.Renderer) {
	if err := render.Render(w, r, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
