package main

import (
	"github.com/mkuznets/classbox/pkg/api"
	"github.com/mkuznets/classbox/pkg/opts"
	"github.com/mkuznets/classbox/pkg/utils"
	"log"
)

// APICommand with command line flags and env
type APICommand struct {
	Addr   string       `long:"addr" env:"ADDR" description:"HTTP service address" default:"127.0.0.1:8080"`
	WebURL string       `long:"web-url" env:"WEB_URL" description:"url to website" required:"true"`
	DB     *opts.DB     `group:"PostgreSQL" namespace:"db" env-namespace:"DB"`
	Github *opts.Github `group:"github" namespace:"github" env-namespace:"GITHUB"`
	AWS    *opts.AWS    `group:"AWS" namespace:"aws" env-namespace:"AWS"`
}

// Execute is the entry point for "api" command, called by flag parser
func (s *APICommand) Execute(args []string) error {
	db, err := s.DB.GetPool()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	log.Print("[INFO] connected to DB")

	server := api.Server{
		Addr: s.Addr,
		API: api.API{
			DB:          db,
			OAuth:       s.Github.OAuth.Config(),
			App:         s.Github.App,
			AWS:         s.AWS,
			RandomState: utils.RandomString(32),
			WebUrl:      s.WebURL,
		},
	}
	server.Start()
	return nil
}

func init() {
	var runCommand APICommand
	_, err := parser.AddCommand(
		"api",
		"start webserver",
		"start webserver",
		&runCommand)
	if err != nil {
		panic(err)
	}
}
