package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
)

func checkHookSignature(secret, signature string, body []byte) (bool, error) {
	requestMac, err := hex.DecodeString(strings.TrimPrefix(signature, "sha1="))
	if err != nil {
		return false, errors.WithStack(err)
	}
	mac := hmac.New(sha1.New, []byte(secret))
	if _, err := mac.Write(body); err != nil {
		return false, errors.WithStack(err)
	}
	return hmac.Equal(requestMac, mac.Sum(nil)), nil
}

func hookValidator(secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				E.Handle(w, r, errors.Wrap(err, "could not read request body"))
				return
			}
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			//noinspection GoUnhandledErrorResult
			defer r.Body.Close()

			signature := r.Header.Get("X-Hub-Signature")
			ok, err := checkHookSignature(secret, signature, body)
			if err != nil {
				E.Handle(w, r, errors.Wrap(err, "could not check signature"))
				return
			}
			if !ok {
				E.SendError(w, r, nil, http.StatusUnauthorized, "webhook signature is invalid")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
