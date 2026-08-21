package general

import (
	"flasher/commandline"
	"flasher/core"
)

const (
	HawkTuahBanner = `
	   --------------------------------------------
		▌ ▌       ▌ ▀▛▘     ▌  ▛▀▘▜       ▌        
		▙▄▌▝▀▖▌  ▌▌▗▘▌▌ ▌▝▀▖▛▀▖▙▄ ▐ ▝▀▖▞▀▘▛▀▖▞▀▖▙▀▖
		▌ ▌▞▀▌▐▐▐ ▛▚ ▌▌ ▌▞▀▌▌ ▌▌  ▐ ▞▀▌▝▀▖▌ ▌▛▀ ▌  
		▘ ▘▝▀▘ ▘▘ ▘ ▘▘▝▀▘▝▀▘▘ ▘▘   ▘▝▀▘▀▀ ▘ ▘▝▀▘▘
	   -------------------------------------------- 
	`
)

func RegisterVersion(parser *commandline.CommandParser) error {
	if parser == nil {
		return core.Errorf("command parser is nil")
	}

	command, err := parser.CreateCommand("version", "Display HawkTuahFlasher's binary version")
	if err != nil {
		return err
	}

	command.Category = "general"
	command.Aliases[0] = "v"
	command.Arguments = nil

	command.Run = RunVersion

	return parser.RegisterCommand(command)
}

func RunVersion(context *commandline.Context, arguments []string) error {
	context.PrintFormat("%s", HawkTuahBanner)
	context.PrintLine()
	context.PrintFormat("%s - command line interface %s", commandline.Name, commandline.Version);

	return nil
}