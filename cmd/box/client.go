package main

import (
	"context"
	"github.com/mkuznets/classbox/pkg/client"
)

// ClientCommand with command line flags and env
type ClientCommand struct {
	ApiURL string `long:"api-url" env:"API_URL" description:"base API URL" required:"true"`
}

// Execute is the entry point for "server" command, called by flag parser
func (s *ClientCommand) Execute(args []string) error {
	ctx := context.Background()
	client.New(ctx, s.ApiURL)
	<-ctx.Done()
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
