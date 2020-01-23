package opts

import "golang.org/x/oauth2"

// Github contains settings for Github apps
type Github struct {
	OAuth OAuth `group:"Github OAuth" namespace:"oauth" env-namespace:"OAUTH"`
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
