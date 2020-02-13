package main

import (
	"context"
	"fmt"
	"github.com/mkuznets/classbox/pkg/opts"
	"github.com/mkuznets/classbox/pkg/runner"
	"net/http"
)

// RunnerCommand with command line flags and env
type RunnerCommand struct {
	Env     *opts.Env       `group:"Environment" namespace:"env" env-namespace:"ENV"`
	ApiURL  string          `long:"api-url" env:"API_URL" description:"base API URL" required:"true"`
	DataDir string          `long:"data-dir" env:"DATA_DIR" description:"exposed data directory" required:"true"`
	WebURL  string          `long:"web-url" env:"WEB_URL" description:"url to website" required:"true"`
	DocsURL string          `long:"docs-url" env:"DOCS_URL" description:"url to generated docs" required:"true"`
	Jwt     *opts.JwtClient `group:"JWT" namespace:"jwt" env-namespace:"JWT"`
	Sentry  *opts.Sentry    `group:"Sentry" namespace:"sentry" env-namespace:"SENTRY"`
	Debug   bool            `long:"debug" description:"show debug info" required:"false"`
}

// Execute is the entry point for "server" command, called by flag parser
func (s *RunnerCommand) Execute(args []string) error {
	ctx := context.Background()

	token, err := s.Jwt.Token()
	if err != nil {
		return err
	}
	if s.Debug {
		fmt.Printf("JWT token: %s\n", token.AccessToken)
		return nil
	}

	cl := &runner.Runner{
		Ctx:     ctx,
		Env:     s.Env,
		Sentry:  s.Sentry,
		Http:    &http.Client{},
		Jwt:     s.Jwt,
		ApiURL:  s.ApiURL,
		WebURL:  s.WebURL,
		DocsURL: s.DocsURL,
		DataDir: s.DataDir,
	}
	cl.Do()

	return nil
}

func init() {
	var cmd RunnerCommand
	_, err := parser.AddCommand(
		"runner",
		"start runner",
		"start runner",
		&cmd)
	if err != nil {
		panic(err)
	}
}
