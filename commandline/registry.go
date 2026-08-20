package commandline

import "sync"

type CommandRegistration func(*CommandParser) error

var (
	registryMutex   sync.RWMutex
	registry      []CommandRegistration
)

func RegisterCommand(registration CommandRegistration) {
	if registration == nil {
		return
	}

	registryMutex.Lock()
	defer registryMutex.Unlock()

	registry = append(registry, registration)
}

func RegisterCommands(parser *CommandParser) error {
	if parser == nil {
		return nil
	}

	registryMutex.RLock()
	registrations := append([]CommandRegistration(nil), registry...)
	registryMutex.RUnlock()

	for _, registration := range registrations {
		if err := registration(parser); err != nil {
			return err
		}
	}

	return nil
}