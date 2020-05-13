package opts

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type JwtServer struct {
	PublicKey string `long:"public-key" env:"PUBLIC_KEY" description:"RSA public key" required:"true"`
}

func (j *JwtServer) Key(token *jwt.Token) (interface{}, error) {
	pemRaw, err := base64.StdEncoding.DecodeString(j.PublicKey)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pemRaw)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM: `PUBLIC KEY` expected")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	if _, ok := key.(*rsa.PublicKey); !ok {
		return nil, errors.New("RSA public key expected")
	}
	return key, nil
}

type JwtClient struct {
	PrivateKey string `long:"private-key" env:"PRIVATE_KEY" description:"RSA private key" required:"true"`
}

func (j *JwtClient) Token() (*oauth2.Token, error) {
	pemRaw, err := base64.StdEncoding.DecodeString(j.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode private key")
	}
	block, _ := pem.Decode(pemRaw)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM: `PRIVATE KEY` expected")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse private key")
	}
	if _, ok := key.(*rsa.PrivateKey); !ok {
		return nil, errors.New("RSA private key expected")
	}
	jwtEncoder := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
	})
	token, err := jwtEncoder.SignedString(key)
	if err != nil {
		return nil, errors.Wrap(err, "could not create signed token")
	}
	return &oauth2.Token{AccessToken: token, TokenType: "bearer"}, nil
}
