package hawktuah

import (
	"flasher/core"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Registry struct {
	ReadWriteMutex sync.RWMutex
	definitions    map[string] *Definition
}

func NewRegistry() *Registry {
	return &Registry{
		definitions: make(map[string]*Definition),
	}
}

func (registry *Registry) Add(definition *Definition) error {
	if registry == nil {
		return core.Errorf("registry is nil")
	}

	if definition == nil {
		return core.Errorf("definition is nil")
	}

	if err := definition.Validate(); err != nil {
		return core.Errorf("invalid definition: %w", err)
	}

	if definition.Name == "" {
		return core.Errorf("definition name is empty")
	}

	registry.ReadWriteMutex.Lock()
	defer registry.ReadWriteMutex.Unlock()

	if _, exists := registry.definitions[definition.Name]; exists {
		return core.Errorf("definition %q is already registered", definition.Name)
	}

	registry.definitions[definition.Name] = definition

	return nil
}

func (registry *Registry) Remove(name string) bool {
	if registry == nil {
		return false
	}

	registry.ReadWriteMutex.Lock()
	defer registry.ReadWriteMutex.Unlock()

	if _, exists := registry.definitions[name]; !exists {
		return false
	}

	delete(registry.definitions, name)

	return true
}

func (registry *Registry) Get(name string) *Definition {
	if registry == nil {
		return nil
	}

	registry.ReadWriteMutex.RLock()
	defer registry.ReadWriteMutex.RUnlock()

	return registry.definitions[name]
}

func (registry *Registry) FindVID(vid uint16) *Definition {
	if registry == nil {
		return nil
	}

	registry.ReadWriteMutex.RLock()
	defer registry.ReadWriteMutex.RUnlock()

	for _, definition := range registry.definitions {
		if definition.FindVID(vid) != nil {
			return definition
		}
	}

	return nil
}

func (registry *Registry) Load(path string) (*Definition, error) {
	definition, err := Load(path)
	if err != nil {
		return nil, err
	}

	if err := registry.Add(definition); err != nil {
		return nil, err
	}

	return definition, nil
}

func (registry *Registry) Definitions() []*Definition {
	if registry == nil {
		return nil
	}

	registry.ReadWriteMutex.RLock()
	defer registry.ReadWriteMutex.RUnlock()

	definitions := make([]*Definition, 0, len(registry.definitions))

	for _, definition := range registry.definitions {
		definitions = append(definitions, definition)
	}

	return definitions
}

func (registry *Registry) Len() int {
	if registry == nil {
		return 0
	}

	registry.ReadWriteMutex.RLock()
	defer registry.ReadWriteMutex.RUnlock()

	return len(registry.definitions)
}

func (registry *Registry) Clear() {
	if registry == nil {
		return
	}

	registry.ReadWriteMutex.Lock()
	defer registry.ReadWriteMutex.Unlock()

	clear(registry.definitions)
}

func (registry *Registry) LoadDirectory(path string) error {
	if registry == nil {
		return core.Errorf("registry is nil")
	}

	if path == "" {
		return core.Errorf("definition directory is empty")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return core.Errorf("read definition directory %q: %w", path, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.ToLower(filepath.Ext(entry.Name())) != ".hawktuah" {
			continue
		}

		filePath := filepath.Join(path, entry.Name())

		if _, err := registry.Load(filePath); err != nil {
			return core.Errorf("load definition %q: %w", filePath, err)
		}
	}

	return nil
}