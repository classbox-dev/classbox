package errors

import (
	"net/http"

	"github.com/go-chi/render"
)

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
