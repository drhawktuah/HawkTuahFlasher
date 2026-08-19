package commandline

import (
	"fmt"
	"strings"
)

const (
	MaxPrefixSize   = 4
	MaxPrefixLength = 6
)

type Prefix struct {
	Value string
}

type Argument struct {
	Name        string
	Description string
	Required    bool
}

type Option struct {
	Name        string
	Aliases     []string
	Description string

	Required bool
	HasValue bool
	Default  string
}

type Flag struct {
	Name        string
	Aliases     []string
	Description string

	Default bool
}

type CommandArguments struct {
	Arguments []Argument
	Options   []Option
	Flags     []Flag
}

type ParsedArguments struct {
	Raw         []string
	Positionals []string
	Options     map[string]string
	Flags       map[string]bool
}

type ParsedCommand struct {
	Command   *Command
	Arguments ParsedArguments
}

type CommandParser struct {
	Commands *Commands
	Prefixes []Prefix
}

func NewParser(commands *Commands) *CommandParser {
	return &CommandParser{
		Commands: commands,
	}
}

func (parser *CommandParser) CreateCommand(name string, description string) (*Command, error) {
	if parser == nil {
		return nil, fmt.Errorf("command parser is nil")
	}

	if parser.Commands == nil {
		return nil, fmt.Errorf("command parser has no command registry")
	}

	name = strings.TrimSpace(name)

	if name == "" {
		return nil, fmt.Errorf("command name cannot be empty")
	}

	if len(name) > MaxCommandNameLength {
		return nil, fmt.Errorf("command name %q exceeds maximum length of %d", name, MaxCommandNameLength)
	}

	command := Command{
		Name:        name,
		Description: description,
	}

	if err := parser.Commands.RegisterCommand(command); err != nil {
		return nil, err
	}

	index := parser.Commands.findAvailableIndex(name)

	if index < 0 {
		return nil, fmt.Errorf("command %q was registered but could not be located", name)
	}

	return &parser.Commands.Available[index], nil
}

func (parser *CommandParser) Parse(arguments []string) (*ParsedCommand, error) {
	if parser == nil {
		return nil, fmt.Errorf("command parser is nil")
	}

	if parser.Commands == nil {
		return nil, fmt.Errorf("command parser has no command registry")
	}

	if len(arguments) == 0 {
		return nil, fmt.Errorf("no command specified")
	}

	commandName := strings.TrimSpace(arguments[0])

	if commandName == "" {
		return nil, fmt.Errorf("command name cannot be empty")
	}

	command := parser.Commands.Find(commandName)

	if command == nil {
		return nil, fmt.Errorf("unknown command %q", commandName)
	}

	if !parser.Commands.IsAvailable(command.Name) {
		return nil, fmt.Errorf("command %q is unavailable", command.Name)
	}

	parsed, err := parser.parseArguments(command, arguments[1:])

	if err != nil {
		return nil, err
	}

	return &ParsedCommand{
		Command:   command,
		Arguments: *parsed,
	}, nil
}

func (parser *CommandParser) UsePrefixes(prefixes ...string) (*CommandParser, error) {
	if len(prefixes) > MaxPrefixSize {
		return nil, fmt.Errorf("max amount of prefixes allowed is %d", MaxPrefixSize)
	}

	for _, value := range prefixes {
		value = strings.TrimSpace(value)

		if value == "" {
			continue
		}

		if len(value) > MaxPrefixLength {
			return nil, fmt.Errorf("prefix length for '%q' is %d. expected length of at most %d", value, len(value), MaxPrefixLength)
		}

		exists := false

		for _, prefix := range parser.Prefixes {
			if prefix.Value == value {
				exists = true
				break
			}
		}

		if exists {
			continue
		}

		if len(parser.Prefixes) >= MaxPrefixSize {
			return nil, fmt.Errorf("max amount of prefixes allowed is %d", MaxPrefixSize)
		}

		parser.Prefixes = append(parser.Prefixes, Prefix{
			Value: value,
		})
	}

	if prefixesLength := len(parser.Prefixes); prefixesLength < 1 {
		prefix_one := Prefix {
			Value: "--",
		}

		prefix_two := Prefix {
			Value: "-",
		}

		prefix_three := Prefix {
			Value: "/",
		}

		parser.Prefixes = append(parser.Prefixes, prefix_one, prefix_two, prefix_three)
	}

	return parser, nil
}

func (parser *CommandParser) parseArguments(command *Command, arguments []string) (*ParsedArguments, error) {
	result := &ParsedArguments{
		Raw:         append([]string(nil), arguments...),
		Positionals: make([]string, 0),
		Options:     make(map[string]string),
		Flags:       make(map[string]bool),
	}

	for _, flag := range command.Flags {
		result.Flags[flag.Name] = flag.Default
	}

	for _, option := range command.Options {
		if option.Default != "" {
			result.Options[option.Name] = option.Default
		}
	}

	for index := 0; index < len(arguments); index++ {
		value := arguments[index]
		_, remainder, prefixed := parser.matchPrefix(value)

		if !prefixed {
			result.Positionals = append(result.Positionals, value)
			continue
		}

		if remainder == "" {
			result.Positionals = append(result.Positionals, arguments[index + 1:]...)
			break
		}

		next, err := parser.parsePrefixed(command, remainder, arguments, index, result)

		if err != nil {
			return nil, err
		}

		index = next
	}

	if err := parser.validateArguments(command, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (parser *CommandParser) matchPrefix(value string) (Prefix, string, bool) {
	var matched Prefix
	var remainder string

	for _, prefix := range parser.Prefixes {
		if !strings.HasPrefix(value, prefix.Value) {
			continue
		}

		if len(prefix.Value) <= len(matched.Value) {
			continue
		}

		matched = prefix
		remainder = strings.TrimPrefix(value, prefix.Value)
	}

	if matched.Value == "" {
		return Prefix{}, "", false
	}

	return matched, remainder, true
}

func (parser *CommandParser) parsePrefixed(command *Command, value string, arguments []string, index int, result *ParsedArguments) (int, error) {
	name := value
	optionValue := ""

	if equals := strings.IndexByte(value, '='); equals >= 0 {
		name = value[:equals]
		optionValue = value[equals + 1:]
	}

	if flag, found := findFlag(command, name); found {
		if optionValue != "" {
			return index, fmt.Errorf("flag %q does not accept a value", name)
		}

		result.Flags[flag.Name] = true
		return index, nil
	}

	option, found := findOption(command, name)

	if !found {
		return index, fmt.Errorf("unknown option %q", name)
	}

	if !option.HasValue {
		if optionValue != "" {
			return index, fmt.Errorf("option %q does not accept a value", name)
		}

		result.Options[option.Name] = "true"
		return index, nil
	}

	if optionValue == "" {
		if index + 1 >= len(arguments) {
			return index, fmt.Errorf("option %q requires a value", name)
		}

		index++
		optionValue = arguments[index]
	}

	result.Options[option.Name] = optionValue
	return index, nil
}

func findOption(command *Command, name string) (Option, bool) {
	name = strings.TrimSpace(name)

	if name == "" {
		return Option{}, false
	}

	for _, option := range command.Options {
		if option.Name == name {
			return option, true
		}

		for _, alias := range option.Aliases {
			if strings.TrimSpace(alias) == name {
				return option, true
			}
		}
	}

	return Option{}, false
}

func findFlag(command *Command, name string) (Flag, bool) {
	name = strings.TrimSpace(name)

	if name == "" {
		return Flag{}, false
	}

	for _, flag := range command.Flags {
		if flag.Name == name {
			return flag, true
		}

		for _, alias := range flag.Aliases {
			if strings.TrimSpace(alias) == name {
				return flag, true
			}
		}
	}

	return Flag{}, false
}

func (parser *CommandParser) validateArguments(command *Command, result *ParsedArguments) error {
	requiredArguments := 0

	for _, argument := range command.Arguments {
		if argument.Required {
			requiredArguments++
		}
	}

	if len(result.Positionals) < requiredArguments {
		return fmt.Errorf("command %q requires at least %d positional argument(s)", command.Name, requiredArguments)
	}

	for _, option := range command.Options {
		if !option.Required {
			continue
		}

		if _, exists := result.Options[option.Name]; !exists {
			return fmt.Errorf("required option %q is missing", option.Name)
		}
	}

	return nil
}

func (parser *CommandParser) Execute(context *Context, arguments []string) error {
	parsed, err := parser.Parse(arguments)
	if err != nil {
		return err
	}

	return parsed.Command.Run(context, parsed.Arguments.Raw)
}
