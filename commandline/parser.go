package commandline

import (
	"flasher/core"
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

func (parser *CommandParser) BuildUsage(command *Command) string {
	if parser == nil || command == nil {
		return ""
	}

	var builder strings.Builder

	builder.WriteString(command.Name)

	for _, argument := range command.Arguments {
		if argument.Name == "" {
			continue
		}

		if argument.Required {
			core.Fprintf(&builder, " <%s>", argument.Name)
		} else {
			core.Fprintf(&builder, " [%s]", argument.Name)
		}
	}

	for _, option := range command.Options {
		name := option.Name

		if name == "" {
			continue
		}

		prefix := parser.optionPrefix()

		if option.HasValue {
			name += "=<value>"
		}

		if option.Required {
			core.Fprintf(&builder, " %s%s", prefix, name)
		} else {
			core.Fprintf(&builder, " [%s%s]", prefix, name)
		}
	}

	for _, flag := range command.Flags {
		if flag.Name == "" {
			continue
		}

		prefix := parser.optionPrefix()

		if flag.Default {
			continue
		}

		core.Fprintf(&builder, " [%s%s]", prefix, flag.Name)
	}

	return builder.String()
}

func NewParser(commands *Commands) *CommandParser {
	return &CommandParser{
		Commands: commands,
	}
}

func (parser *CommandParser) CreateCommand(name string, description string) (*Command, error) {
	if parser == nil {
		return nil, core.Errorf("command parser is nil")
	}

	if parser.Commands == nil {
		return nil, core.Errorf("command parser has no command registry")
	}

	name = strings.TrimSpace(name)

	if name == "" {
		return nil, core.Errorf("command name cannot be empty")
	}

	if len(name) > MaxCommandNameLength {
		return nil, core.Errorf("command name '%q' exceeds maximum length of '%d'", name, MaxCommandNameLength)
	}

	return &Command{
		Name:        name,
		Description: description,
	}, nil
}

func (parser *CommandParser) RegisterCommand(command *Command) error {
	if parser == nil {
		return core.Errorf("command parser is nil")
	}

	if parser.Commands == nil {
		return core.Errorf("command parser has no command registry")
	}

	if command == nil {
		return core.Errorf("command is nil")
	}

	command.Name = strings.TrimSpace(command.Name)
	return parser.Commands.RegisterCommand(*command)
}

func (parser *CommandParser) Parse(arguments []string) (*ParsedCommand, error) {
	if parser == nil {
		return nil, core.Errorf("command parser is nil")
	}

	if parser.Commands == nil {
		return nil, core.Errorf("command parser has no command registry")
	}

	if len(arguments) == 0 {
		return nil, core.Errorf("no command specified")
	}

	commandName := strings.TrimSpace(arguments[0])

	if commandName == "" {
		return nil, core.Errorf("command name cannot be empty")
	}

	command := parser.Commands.Find(commandName)

	if command == nil {
		return nil, core.Errorf("unknown command %q", commandName)
	}

	if !parser.Commands.IsAvailable(command.Name) {
		return nil, core.Errorf("command '%q' is unavailable", command.Name)
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
		return nil, core.Errorf("max amount of prefixes allowed is %d", MaxPrefixSize)
	}

	for _, value := range prefixes {
		value = strings.TrimSpace(value)

		if value == "" {
			continue
		}

		if len(value) > MaxPrefixLength {
			return nil, core.Errorf("prefix length for '%q' is %d. expected length of at most %d", value, len(value), MaxPrefixLength)
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
			return nil, core.Errorf("max amount of prefixes allowed is %d", MaxPrefixSize)
		}

		parser.Prefixes = append(parser.Prefixes, Prefix{
			Value: value,
		})
	}

	if prefixesLength := len(parser.Prefixes); prefixesLength < 1 {
		prefix_one := Prefix{
			Value: "--",
		}

		prefix_two := Prefix{
			Value: "-",
		}

		prefix_three := Prefix{
			Value: "/",
		}

		parser.Prefixes = append(parser.Prefixes, prefix_one, prefix_two, prefix_three)
	}

	return parser, nil
}

func (parser *CommandParser) optionPrefix() string {
	if parser == nil {
		return "--"
	}

	var longest string

	for _, prefix := range parser.Prefixes {
		value := strings.TrimSpace(prefix.Value)

		if len(value) > len(longest) {
			longest = value
		}
	}

	if longest == "" {
		return "--"
	}

	return longest
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
			result.Positionals = append(result.Positionals, arguments[index+1:]...)
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
		optionValue = value[equals+1:]
	}

	if flag, found := findFlag(command, name); found {
		if optionValue != "" {
			return index, core.Errorf("flag %q does not accept a value", name)
		}

		result.Flags[flag.Name] = true
		return index, nil
	}

	option, found := findOption(command, name)

	if !found {
		return index, core.Errorf("unknown option %q", name)
	}

	if !option.HasValue {
		if optionValue != "" {
			return index, core.Errorf("option %q does not accept a value", name)
		}

		result.Options[option.Name] = "true"
		return index, nil
	}

	if optionValue == "" {
		if index+1 >= len(arguments) {
			return index, core.Errorf("option %q requires a value", name)
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
		return core.Errorf("command %q requires at least %d positional argument(s)", command.Name, requiredArguments)
	}

	for _, option := range command.Options {
		if !option.Required {
			continue
		}

		if _, exists := result.Options[option.Name]; !exists {
			return core.Errorf("required option %q is missing", option.Name)
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
