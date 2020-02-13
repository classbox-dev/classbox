package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	E "github.com/mkuznets/classbox/pkg/api/errors"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var expirationLimit = 5 * time.Minute

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

func jwtValidator(keyFunc func(token *jwt.Token) (interface{}, error)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var tokenString string
			authHeader := r.Header.Get("Authorization")

			if n, err := fmt.Sscanf(authHeader, "Bearer %s", &tokenString); err != nil || n != 1 {
				E.SendError(w, r, nil, http.StatusUnauthorized, "valid auth header is required")
				return
			}
			var claims jwt.StandardClaims
			_, err := jwt.ParseWithClaims(tokenString, &claims, keyFunc)
			if err != nil {
				E.SendError(w, r, err, http.StatusUnauthorized, "invalid token: "+err.Error())
				return
			}
			delta1 := time.Unix(claims.ExpiresAt, 0).Sub(time.Now())
			delta2 := time.Duration(claims.ExpiresAt-claims.IssuedAt) * time.Second
			if delta1 > expirationLimit || delta2 > expirationLimit {
				E.SendError(w, r, nil, http.StatusUnauthorized, "expiration time cannot exceed 5 minutes")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
