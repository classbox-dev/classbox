package main

import (
	"github.com/mkuznets/classbox/pkg/api"
	"log"
)

// APICommand with command line flags and env
type APICommand struct {
	DB DBOptions `group:"PostgreSQL settings" namespace:"db" env-namespace:"DB"`
}

// Execute is the entry point for "api" command, called by flag parser
func (s *APICommand) Execute(args []string) error {
	db, err := s.DB.GetPool()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	log.Print("[INFO] connected to DB")

	server := api.Server{
		Port: 8080,
		API: api.API{
			DB: db,
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
