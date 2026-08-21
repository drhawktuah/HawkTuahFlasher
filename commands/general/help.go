package general

import (
	"sort"
	"strings"

	"flasher/commandline"
	"flasher/core"
)

const (
	TreeFirstCategory  = "╭─ "
	TreeMiddleCategory = "├─ "
	TreeLastCategory   = "╰─ "

	TreeMiddleCommand = "├─ "
	TreeLastCommand   = "╰─ "

	TreeVertical  = "│  "
	TreeEmpty     = "   "
	TreeSeparator = "┊"
)

type CommandGroup struct {
	Name     string
	Commands []commandline.Command
}

func RegisterHelp(parser *commandline.CommandParser) error {
	if parser == nil {
		return core.Errorf("command parser is nil")
	}

	command, err := parser.CreateCommand(
		"help",
		"Display command help",
	)
	if err != nil {
		return err
	}

	command.Category = "general"
	command.Aliases[0] = "h"
	command.Aliases[1] = "?"

	command.Arguments = []commandline.Argument{
		{
			Name:        "command",
			Description: "Command to display help for",
			Required:    false,
		},
	}

	command.Run = RunHelp

	return parser.RegisterCommand(command)
}

func RunHelp(context *commandline.Context, arguments []string) error {
	if context == nil {
		return core.Errorf("command context is nil")
	}

	if context.Commands == nil {
		return core.Errorf("command registry is unavailable")
	}

	if context.Parser == nil {
		return core.Errorf("command parser is unavailable")
	}

	if len(arguments) > 0 {
		return printCommandHelp(context, context.Parser, context.Commands, arguments[0])
	}

	printCommandTree(context, context.Commands)
	return nil
}

func printCommandTree(context *commandline.Context, commands *commandline.Commands) {
	context.PrintFormat("%s - command line interface %s", commandline.Name, commandline.Version)

	context.PrintLine()

	context.PrintLine("Usage:")
	context.PrintFormat("  %s <command> [arguments]\n", commandline.Name)

	context.PrintLine()
	context.PrintLine("Commands:")
	context.PrintLine()

	groups := groupCommands(commands)

	if len(groups) == 0 {
		context.PrintLine("No commands available.")
		return
	}

	for index, group := range groups {
		printCommandGroup(context, group, index, len(groups))
	}
}

func printCommandGroup(context *commandline.Context, group CommandGroup, index int, total int) {
	if context == nil || len(group.Commands) == 0 {
		return
	}

	lastCategory := index == total-1
	categoryPrefix := TreeMiddleCategory

	switch {
		case total == 1:
			categoryPrefix = TreeFirstCategory

		case index == 0:
			categoryPrefix = TreeFirstCategory

		case lastCategory:
			categoryPrefix = TreeLastCategory
	}

	context.PrintLine(categoryPrefix + group.Name)

	for commandIndex, command := range group.Commands {
		lastCommand := commandIndex == len(group.Commands)-1
		commandPrefix := TreeMiddleCommand

		if lastCommand {
			commandPrefix = TreeLastCommand
		}

		continuation := TreeVertical

		if lastCategory && total > 1 {
			continuation = TreeEmpty
		}

		context.PrintFormat("%s%s%-16s", continuation, commandPrefix, command.Name)

		if command.Description != "" {
			context.PrintFormat(" %s", command.Description)
		}

		context.PrintLine()
	}

	if lastCategory {
		context.PrintLine(TreeLastCategory)
	} else {
		context.PrintLine(TreeSeparator)
	}
}

func groupCommands(commands *commandline.Commands) []CommandGroup {
	if commands == nil {
		return nil
	}

	groups := make(map[string][]commandline.Command)

	for index := 0; index < commands.AvailableLength; index++ {
		command := commands.Available[index]
		category := strings.TrimSpace(command.Category)

		if category == "" {
			category = "system"
		}

		groups[category] = append(groups[category], command)
	}

	result := make([]CommandGroup, 0, len(groups))

	for category, commands := range groups {
		sort.Slice(commands, func(i, j int) bool {
			return commands[i].Name < commands[j].Name
		})

		result = append(result, CommandGroup{
			Name:     category,
			Commands: commands,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

func printCommandHelp(context *commandline.Context, parser *commandline.CommandParser, commands *commandline.Commands, name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return core.Errorf("command name cannot be empty")
	}

	command := commands.Find(name)

	if command == nil {
		return core.Errorf("unknown command %q", name)
	}

	context.PrintLine()

	if command.Category != "" {
		context.PrintFormat("%s/%s\n", command.Category, command.Name)
	} else {
		context.PrintLine(command.Name)
	}

	if command.Description != "" {
		context.PrintLine()
		context.PrintFormat("  %s\n", command.Description)
	}

	usage := parser.BuildUsage(command)

	if usage != "" {
		context.PrintLine()
		context.PrintLine("Usage:")
		context.PrintFormat("  %s %s\n", commandline.Name, usage)
	}

	printAliases(context, command)
	printArguments(context, command)
	printOptions(context, command)
	printFlags(context, command)
	printDependencies(context, command)

	return nil
}

func printAliases(context *commandline.Context, command *commandline.Command) {
	aliases := make([]string, 0)

	for _, alias := range command.Aliases {
		alias = strings.TrimSpace(alias)

		if alias == "" {
			continue
		}

		aliases = append(aliases, alias)
	}

	if len(aliases) == 0 {
		return
	}

	context.PrintLine()
	context.PrintLine("Aliases:", strings.Join(aliases, ", "))
}

func printArguments(context *commandline.Context, command *commandline.Command) {
	if len(command.Arguments) == 0 {
		return
	}

	context.PrintLine()
	context.PrintLine("Arguments:")

	for index, argument := range command.Arguments {
		prefix := TreeMiddleCommand

		if index == len(command.Arguments)-1 {
			prefix = TreeLastCommand
		}

		required := ""

		if argument.Required {
			required = " (required)"
		}

		context.PrintFormat("  %s%-16s %s%s\n", prefix, argument.Name, argument.Description, required)
	}
}

func printOptions(context *commandline.Context, command *commandline.Command) {
	if len(command.Options) == 0 {
		return
	}

	context.PrintLine()
	context.PrintLine("Options:")

	for index, option := range command.Options {
		prefix := TreeMiddleCommand

		if index == len(command.Options)-1 {
			prefix = TreeLastCommand
		}

		required := ""

		if option.Required {
			required = " (required)"
		}

		value := ""

		if option.HasValue {
			value = " <value>"
		}

		context.PrintFormat("  %s%-16s%s %s%s\n", prefix, option.Name, value, option.Description, required)
	}
}

func printFlags(context *commandline.Context, command *commandline.Command) {
	if len(command.Flags) == 0 {
		return
	}

	context.PrintLine()
	context.PrintLine("Flags:")

	for index, flag := range command.Flags {
		prefix := TreeMiddleCommand

		if index == len(command.Flags)-1 {
			prefix = TreeLastCommand
		}

		context.PrintFormat("  %s%-16s %s\n", prefix, flag.Name, flag.Description)
	}
}

func printDependencies(context *commandline.Context, command *commandline.Command) {
	if len(command.Dependencies) == 0 {
		return
	}

	context.PrintLine()
	context.PrintLine("Dependencies:")

	statuses := command.DependencyStatus

	if len(statuses) == 0 {
		for index, dependency := range command.Dependencies {
			prefix := TreeMiddleCommand

			if index == len(command.Dependencies)-1 {
				prefix = TreeLastCommand
			}

			required := "optional"

			if dependency.Required {
				required = "required"
			}

			context.PrintFormat("  %s? %-14s %s (%s)\n", prefix, dependency.Name, dependency.Description, required)
		}

		return
	}

	for index, status := range statuses {
		prefix := TreeMiddleCommand

		if index == len(statuses)-1 {
			prefix = TreeLastCommand
		}

		state := "✓"

		if !status.Available {
			state = "✗"
		}

		required := "optional"

		if status.Dependency.Required {
			required = "required"
		}

		context.PrintFormat("  %s%s %-14s %s (%s)\n", prefix, state, status.Dependency.Name, status.Dependency.Description, required)

		if status.Error != nil {
			context.PrintFormat("  %s  └─ %s\n", TreeVertical, status.Error)
		}
	}
}
