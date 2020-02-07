package main

import (
	"fmt"
	_ "github.com/mkuznets/classbox/pkg/statik"
	"github.com/rakyll/statik/fs"
	"io/ioutil"
	"log"
)

// WebCommand with command line flags and env
type WebCommand struct {
}

// Execute is the entry point for "api" command, called by flag parser
func (s *WebCommand) Execute(args []string) error {

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	// Access individual files by their paths.
	r, err := statikFS.Open("/templates/index.html")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(contents))

	return nil
}

func init() {
	var runCommand WebCommand
	_, err := parser.AddCommand(
		"web",
		"",
		"",
		&runCommand)
	if err != nil {
		panic(err)
	}
}
