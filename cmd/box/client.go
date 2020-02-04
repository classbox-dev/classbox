package main

import (
	"context"
	"github.com/mkuznets/classbox/pkg/client"
	"net/http"
)

// ClientCommand with command line flags and env
type ClientCommand struct {
	ApiURL  string            `long:"api-url" env:"API_URL" description:"base API URL" required:"true"`
	// Volumes map[string]string `short:"v" env:"VOLUMES" env-delim:";" description:"artifacts directory" required:"true"`
}

// Execute is the entry point for "server" command, called by flag parser
func (s *ClientCommand) Execute(args []string) error {
	ctx := context.Background()

	cl := &client.Client{
		Ctx:    ctx,
		Http:   &http.Client{},
		ApiURL: s.ApiURL,
	}
	cl.Do()

	return nil
}

func init() {
	var cmd ClientCommand
	_, err := parser.AddCommand(
		"client",
		"start client",
		"start client",
		&cmd)
	if err != nil {
		panic(err)
	}
}
