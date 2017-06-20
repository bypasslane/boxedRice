package main

import (
	"fmt"
	"os"

	goflags "github.com/jessevdk/go-flags" // rename import to `goflags` (file scope) so we can use `var flags` (package scope)
)

// flags
var flags struct {
	Verbose bool `long:"verbose" short:"v" description:"Show verbose debug information"`

	Append struct {
		BoxPath    []string `long:"box-path" short:"b" description:"Box path(s) to use. Bypasses code parsing to enable simpler use of boxes. Ignores import-path. Specify multiple times for more box paths to append" required:"true"`
		Executable string   `long:"exec" description:"Executable to append" required:"true"`
	} `command:"append"`
}

// flags parser
var flagsParser *goflags.Parser

// initFlags parses the given flags.
// when the user asks for help (-h or --help): the application exists with status 0
// when unexpected flags is given: the application exits with status 1
func parseArguments() {
	// create flags parser in global var, for flagsParser.Active.Name (operation)
	flagsParser = goflags.NewParser(&flags, goflags.Default)

	// parse flags
	args, err := flagsParser.Parse()
	if err != nil {
		// assert the err to be a flags.Error
		flagError := err.(*goflags.Error)
		if flagError.Type == goflags.ErrHelp {
			// user asked for help on flags.
			// program can exit successfully
			os.Exit(0)
		}
		if flagError.Type == goflags.ErrUnknownFlag {
			fmt.Println("Use --help to view available options.")
			os.Exit(1)
		}
		if flagError.Type == goflags.ErrRequired {
			os.Exit(1)
		}
		fmt.Printf("Error parsing flags: %s\n", err)
		os.Exit(1)
	}

	// error on left-over arguments
	if len(args) > 0 {
		fmt.Printf("Unexpected arguments: %s\nUse --help to view available options.", args)
		os.Exit(1)
	}

}
