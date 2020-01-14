package main

import (
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

// GlobalOptions contains cross-command parameters
type GlobalOptions struct{}

var globalOptions GlobalOptions
var parser = flags.NewParser(&globalOptions, flags.Default)

func main() {
	parser.CommandHandler = func(command flags.Commander, args []string) error {
		c := command.(flags.Commander)
		err := c.Execute(args)
		if err != nil {
			log.Printf("[ERROR] failed with %+v", err)
		}
		return err
	}

	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}
