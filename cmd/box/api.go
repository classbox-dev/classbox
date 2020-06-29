package main

import (
	"log"
	"time"

	"github.com/mkuznets/classbox/pkg/api"
	"github.com/mkuznets/classbox/pkg/opts"
	"github.com/mkuznets/classbox/pkg/utils"
)

var maxTime = time.Date(2999, time.December, 31, 0, 0, 0, 0, time.UTC)

// APICommand with command line flags and env
type APICommand struct {
	Env      *opts.Env       `group:"Environment" namespace:"env" env-namespace:"ENV"`
	Addr     string          `long:"addr" env:"ADDR" description:"HTTP service address" default:"127.0.0.1:8080"`
	WebURL   string          `long:"web-url" env:"WEB_URL" description:"url to website" required:"true"`
	Deadline string          `long:"deadline" env:"DEADLINE" description:"submission deadline"`
	DB       *opts.DB        `group:"PostgreSQL" namespace:"db" env-namespace:"DB"`
	Github   *opts.Github    `group:"github" namespace:"github" env-namespace:"GITHUB"`
	AWS      *opts.AWS       `group:"AWS" namespace:"aws" env-namespace:"AWS"`
	Jwt      *opts.JwtServer `group:"JWT" namespace:"jwt" env-namespace:"JWT"`
	Sentry   *opts.Sentry    `group:"Sentry" namespace:"sentry" env-namespace:"SENTRY"`
}

// Execute is the entry point for "api" command, called by flag parser
func (s *APICommand) Execute(args []string) error {
	db, err := s.DB.GetPool()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	log.Print("[INFO] connected to DB")

	deadline := maxTime
	if s.Deadline != "" {
		d, err := time.Parse(time.RFC3339, s.Deadline)
		if err != nil {
			return err
		}
		deadline = d
		log.Printf("[INFO] Submission deadline: %v", deadline)
	}

	server := api.Server{
		Addr:   s.Addr,
		Env:    s.Env,
		Sentry: s.Sentry,
		API: api.API{
			DB:          db,
			OAuth:       s.Github.OAuth,
			App:         s.Github.App,
			AWS:         s.AWS,
			Jwt:         s.Jwt,
			RandomState: utils.RandomString(32),
			WebUrl:      s.WebURL,
			EnvType:     s.Env.Type,
			Deadline:    deadline,
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
