package main

import (
	"github.com/mkuznets/classbox/pkg/api/client"
	_ "github.com/mkuznets/classbox/pkg/statik"
	"github.com/mkuznets/classbox/pkg/web"
)

// WebCommand with command line flags and env
type WebCommand struct {
	Addr    string `long:"addr" env:"ADDR" description:"HTTP service address" default:"127.0.0.1:8082"`
	ApiURL  string `long:"api-url" env:"API_URL" description:"base API URL" required:"true"`
	DocsURL string `long:"docs-url" env:"DOCS_URL" description:"url to generated docs" required:"true"`
}

// Execute is the entry point for "api" command, called by flag parser
func (s *WebCommand) Execute(args []string) error {

	ts, err := web.NewTemplates()
	if err != nil {
		return err
	}

	server := web.Server{
		Addr: s.Addr,
		Web: &web.Web{
			API:       client.New(s.ApiURL),
			Templates: ts,
			DocsURL:   s.DocsURL,
		},
	}
	server.Start()
	return nil
}

func init() {
	var runCommand WebCommand
	_, err := parser.AddCommand("web", "", "",
		&runCommand)
	if err != nil {
		panic(err)
	}
}
