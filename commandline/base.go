package commandline

import (
	stdcontext "context"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

const (
	Name    = "HawkTuahFlasher"
	Version = "Beta-0.1.0"
)

const (
	MaxAliasesLength     = 10
	MaxAliasLength       = 8
	MaxCommandNameLength = 16
	MaxCommandsLength    = 512
)

type Command struct {
	Aliases     [MaxAliasesLength]string
	Name        				  string
	Description                   string
	Category					  string
	Usage       				  string

	Arguments		 []Argument
	Options   		 []Option
	Flags     		 []Flag

	Dependencies     []Dependency
	DependencyStatus []DependencyStatus

	Run func(context *Context, arguments []string) error
}

type Context struct {
	Context 	   stdcontext.Context

	StandardInput  io.Reader
	StandardOutput io.Writer
	StandardError  io.Writer

	Commands 	  *Commands
	Parser 		  *CommandParser

	Verbose        bool
}

type Dependency struct {
	Name        string
	Description string
	Required    bool

	Check func(*Context) error
}

type DependencyStatus struct {
	Dependency Dependency
	Available  bool
	Error      error
}

func NewContext(context stdcontext.Context) *Context {
	if context == nil {
		context = stdcontext.Background()
	}

	return &Context{
		Context: context,

		StandardInput:  os.Stdin,
		StandardOutput: os.Stdout,
		StandardError:  os.Stderr,
	}
}

func (context *Context) PrintLine(arguments ...any) {
	if context == nil || context.StandardOutput == nil {
		return
	}

	fmt.Fprintln(context.StandardOutput, arguments...)
}

func (context *Context) PrintFormat(format string, arguments ...any) {
	if context == nil || context.StandardOutput == nil {
		return
	}

	fmt.Fprintf(context.StandardOutput, format, arguments...)
}

func (context *Context) ErrorLine(arguments ...any) {
	if context == nil || context.StandardOutput == nil {
		return
	}

	fmt.Fprintln(context.StandardError, arguments...)
}

type Commands struct {
	Available   [MaxCommandsLength]Command
	Unavailable [MaxCommandsLength]Command

	AvailableLength   int
	UnavailableLength int
}

func (commands *Commands) IsAvailable(name string) bool {
	if commands == nil {
		return false
	}

	name = strings.TrimSpace(name)

	if name == "" {
		return false
	}

	return commands.findAvailableIndex(name) >= 0
}

func (commands *Commands) removeUnavailable(index int) {
	if index < 0 || index >= commands.UnavailableLength {
		return
	}

	copy(commands.Unavailable[index:], commands.Unavailable[index + 1:commands.UnavailableLength])

	commands.UnavailableLength--
	commands.Unavailable[commands.UnavailableLength] = Command{}
}

func (commands *Commands) removeAvailable(index int) {
	if index < 0 || index >= commands.AvailableLength {
		return
	}

	copy(commands.Available[index:], commands.Available[index + 1:commands.AvailableLength])

	commands.AvailableLength--
	commands.Available[commands.AvailableLength] = Command{}
}

func (commands *Commands) findUnavailableIndex(name string) int {
	name = strings.TrimSpace(name)

	for index := 0; index < commands.UnavailableLength; index++ {
		command := &commands.Unavailable[index]

		if command.Name == name {
			return index
		}

		for _, alias := range command.Aliases {
			if strings.TrimSpace(alias) == name {
				return index
			}
		}
	}

	return -1
}

func (commands *Commands) findAvailableIndex(name string) int {
	name = strings.TrimSpace(name)

	for index := 0; index < commands.AvailableLength; index++ {
		command := &commands.Available[index]

		if command.Name == name {
			return index
		}

		for _, alias := range command.Aliases {
			if strings.TrimSpace(alias) == name {
				return index
			}
		}
	}

	return -1
}

func (commands *Commands) moveAvailableToUnavailable(index int) bool {
	if index < 0 || index >= commands.AvailableLength {
		return false
	}

	if commands.UnavailableLength >= MaxCommandsLength {
		return false
	}

	command := commands.Available[index]

	commands.removeAvailable(index)

	commands.Unavailable[commands.UnavailableLength] = command
	commands.UnavailableLength++

	return true
}

func (commands *Commands) moveUnavailableToAvailable(index int) bool {
	if index < 0 || index >= commands.UnavailableLength {
		return false
	}

	if commands.AvailableLength >= MaxCommandsLength {
		return false
	}

	command := commands.Unavailable[index]

	commands.removeUnavailable(index)

	commands.Available[commands.AvailableLength] = command
	commands.AvailableLength++

	return true
}

func (commands *Commands) dependenciesAvailable(command *Command, context *Context) bool {
	for _, dependency := range command.Dependencies {
		if dependency.Check == nil {
			if dependency.Required {
				return false
			}

			continue
		}

		if err := dependency.Check(context); err != nil {
			if dependency.Required {
				return false
			}
		}
	}

	return true
}

func (commands *Commands) RegisterCommand(command Command) error {
	command.Name = strings.TrimSpace(command.Name)
	command.Category = strings.TrimSpace(command.Category)
	command.Description = strings.TrimSpace(command.Description)
	command.Usage = strings.TrimSpace(command.Usage)

	if command.Name == "" {
		return fmt.Errorf("command name cannot be empty")
	}

	if len(command.Name) > MaxCommandNameLength {
		return fmt.Errorf("command name %q exceeds maximum length of %d", command.Name, MaxCommandNameLength)
	}

	if command.Run == nil {
		return fmt.Errorf("command %q has no run function", command.Name)
	}

	if commands.Exists(command.Name) {
		return fmt.Errorf("command %q is already registered", command.Name)
	}

	aliasCount := 0

	for index := range command.Aliases {
		alias := strings.TrimSpace(command.Aliases[index])

		if alias == "" {
			continue
		}

		if len(alias) > MaxAliasLength {
			return fmt.Errorf("alias '%q' for command '%q' exceeds maximum length of %d", alias, command.Name, MaxAliasLength)
		}

		if alias == command.Name {
			return fmt.Errorf("alias '%q' for command '%q' is identical to the command name", alias, command.Name)
		}

		command.Aliases[index] = alias
		aliasCount++
	}

	for _, alias := range command.Aliases {
		if alias == "" {
			continue
		}

		if commands.Exists(alias) {
			return fmt.Errorf("alias '%q' for command '%q' is already registered", alias, command.Name)
		}
	}

	for index := 0; index < MaxAliasesLength; index++ {
		alias := command.Aliases[index]

		if alias == "" {
			continue
		}

		if slices.Contains(command.Aliases[index + 1:], alias) {
			return fmt.Errorf("alias '%q' is registered more than once for command '%q'", alias, command.Name)
		}
	}

	if aliasCount > MaxAliasesLength {
		return fmt.Errorf("command '%q' exceeds maximum alias count of '%d'", command.Name, MaxAliasesLength)
	}

	if commands.AvailableLength >= MaxCommandsLength {
		return fmt.Errorf("maximum number of available commands (%d) reached", MaxCommandsLength)
	}

	commands.Available[commands.AvailableLength] = command
	commands.AvailableLength++

	return nil
}

func (commands *Commands) UnregisterCommand(command Command) bool {
	if index := commands.findAvailableIndex(command.Name); index >= 0 {
		commands.removeAvailable(index)
		return true
	}

	if index := commands.findUnavailableIndex(command.Name); index >= 0 {
		commands.removeUnavailable(index)
		return true
	}

	return false
}

func (commands *Commands) Exists(name string) bool {
	name = strings.TrimSpace(name)

	if name == "" {
		return false
	}

	return commands.findAvailableIndex(name) >= 0 || commands.findUnavailableIndex(name) >= 0
}

func (commands *Commands) Find(name string) *Command {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil
	}

	if index := commands.findAvailableIndex(name); index >= 0 {
		return &commands.Available[index]
	}

	if index := commands.findUnavailableIndex(name); index >= 0 {
		return &commands.Unavailable[index]
	}

	return nil
}

func (commands *Commands) SetAvailable(name string, available bool) bool {
	name = strings.TrimSpace(name)

	if name == "" {
		return false
	}

	if available {
		index := commands.findUnavailableIndex(name)

		if index < 0 {
			return commands.findAvailableIndex(name) >= 0
		}

		return commands.moveUnavailableToAvailable(index)
	}

	index := commands.findAvailableIndex(name)

	if index < 0 {
		return commands.findUnavailableIndex(name) >= 0
	}

	return commands.moveAvailableToUnavailable(index)
}

func (commands *Commands) CheckDependencies(command *Command, context *Context) bool {
	if command == nil {
		return false
	}

	command.DependencyStatus = command.DependencyStatus[:0]

	available := true

	for _, dependency := range command.Dependencies {
		status := DependencyStatus{
			Dependency: dependency,
			Available:  true,
		}

		if dependency.Check == nil {
			status.Available = false
			status.Error = fmt.Errorf("dependency %q has no check function", dependency.Name)

			if dependency.Required {
				available = false
			}

			command.DependencyStatus = append(command.DependencyStatus, status)
			continue
		}

		if err := dependency.Check(context); err != nil {
			status.Available = false
			status.Error = err

			if dependency.Required {
				available = false
			}
		}

		command.DependencyStatus = append(command.DependencyStatus, status)
	}

	return available
}

func (commands *Commands) CheckAllDependencies(context *Context) {
	if commands == nil {
		return
	}

	for index := 0; index < commands.AvailableLength; index++ {
		command := &commands.Available[index]
		commands.CheckDependencies(command, context)
	}

	for index := 0; index < commands.UnavailableLength; index++ {
		command := &commands.Unavailable[index]
		commands.CheckDependencies(command, context)
	}
}