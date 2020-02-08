package main

import (
	"context"
	"github.com/mkuznets/classbox/pkg/runner"
	"net/http"
)

// RunnerCommand with command line flags and env
type RunnerCommand struct {
	ApiURL  string `long:"api-url" env:"API_URL" description:"base API URL" required:"true"`
	DataDir string `long:"data-dir" env:"DATA_DIR" description:"exposed data directory" required:"true"`
	WebURL  string `long:"web-url" env:"WEB_URL" description:"url to website" required:"true"`
	DocsURL string `long:"docs-url" env:"DOCS_URL" description:"url to generated docs" required:"true"`
}

// Execute is the entry point for "server" command, called by flag parser
func (s *RunnerCommand) Execute(args []string) error {
	ctx := context.Background()

	cl := &runner.Runner{
		Ctx:     ctx,
		Http:    &http.Client{},
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
