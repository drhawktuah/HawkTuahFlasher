package hawktuah

import (
	"flasher/core"
	"os"
)

func Load(path string) (*Definition, error) {
	if path == "" {
		return nil, core.Errorf("definition path is empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, core.Errorf("read definition %q: %w", path, err)
	}

	definition, err := LoadBytes(data)
	if err != nil {
		return nil, core.Errorf("parse definition %q: %w", path, err)
	}

	return definition, nil
}

func LoadBytes(data []byte) (*Definition, error) {
	return LoadString(string(data))
}

func LoadString(source string) (*Definition, error) {
	definition, err := Parse(source)
	if err != nil {
		return nil, core.Errorf("parse definition: %w", err)
	}

	if err := definition.Validate(); err != nil {
		return nil, core.Errorf("validate definition: %w", err)
	}

	return definition, nil
}