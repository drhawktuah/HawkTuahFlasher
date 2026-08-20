package main

import (
	"fmt"
	"os"

	"flasher/commandline"
	general "flasher/commands/general"
)

func main() {
	commands := &commandline.Commands{}

	parser := commandline.NewParser(commands)

	if _, err := parser.UsePrefixes("--", "-", "/"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := general.Register(parser); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	context := commandline.NewContext(nil)
	context.Commands = commands
	context.Parser = parser

	commands.CheckAllDependencies(context)

	arguments := os.Args[1:]

	if len(arguments) == 0 {
		arguments = []string{"help"}
	}

	if err := parser.Execute(context, arguments); err != nil {
		context.ErrorLine(err)
		os.Exit(1)
	}
}