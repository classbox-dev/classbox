package main

import (
	"context"
	"log"
)

// ClientCommand with command line flags and env
type ClientCommand struct {
	DB DBOptions `group:"PostgreSQL settings" namespace:"db" env-namespace:"DB"`
}

// Execute is the entry point for "server" command, called by flag parser
func (s *ClientCommand) Execute(args []string) error {
	db, err := s.DB.GetPool()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	log.Print("[INFO] connected to DB")

	ctx := context.Background()
	New(ctx, "submission", db)
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
