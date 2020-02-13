package opts

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
	"time"
)

// Github contains settings for Github apps
type Github struct {
	OAuth *OAuth `group:"Github OAuth" namespace:"oauth" env-namespace:"OAUTH"`
	App   *App   `group:"Github App" namespace:"app" env-namespace:"APP"`
}

// OAuth contains settings of Github OAuth app
type OAuth struct {
	ClientID     string `long:"client-id" env:"CLIENT_ID" description:"client id" required:"true"`
	ClientSecret string `long:"client-secret" env:"CLIENT_SECRET" description:"client secret" required:"true"`
}

// Config returns `oauth2.Config` for the given settings
func (g *OAuth) Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     g.ClientID,
		ClientSecret: g.ClientSecret,
		Scopes:       []string{"repo"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}
}

// App contains settings of (native) Github App
type App struct {
	ID         string `long:"id" env:"ID" description:"app id" required:"true"`
	Name       string `long:"name" env:"NAME" description:"app name" required:"true"`
	HookSecret string `long:"hook-secret" env:"HOOK_SECRET" required:"true"`
	PrivateKey string `long:"private-key" env:"PRIVATE_KEY" description:"base64-encoded private key in pem format" required:"true"`
}

// Token returns JWT token from the configured app key
func (app *App) Token() (*oauth2.Token, error) {

	pemRaw, err := base64.StdEncoding.DecodeString(app.PrivateKey)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemRaw)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block with private key")
	}

	pkey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	jwtEncoder := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(10 * time.Minute).Unix(),
		Issuer:    app.ID,
	})

	token, err := jwtEncoder.SignedString(pkey)
	if err != nil {
		panic(err)
	}

	return &oauth2.Token{
		AccessToken: token,
		TokenType:   "bearer",
	}, nil
}
