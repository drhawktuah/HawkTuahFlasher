package core

import (
	"os"
	"fmt"
	"strconv"
	"strings"
)

// Base definition for .hawktuah file descriptions
type Definition struct {
	Descriptors Descriptors
	Detect      DetectConfig
	Protocol    ProtocolConfig
	Custom      map[string]string
	CMake       CMakeConfig
	Flash       FlashConfig
}

type Descriptors struct {
	Name         string
	Vendor       string
	Family       string
	Architecture string
	Description  string
}

type DetectConfig struct {
	Transport    string
	VID          []uint16
	PID          []uint16
	Manufacturer string
	Product      string
	Probe        bool
}

type ProtocolConfig struct {
	Name    string
	Version string
}

type CMakeConfig struct {
	Toolchain     string
	Compiler      string
	Architecture  string
	ToolchainFile string
}

type FlashConfig struct {
	Offset   uint64
	Erase    bool
	Verify   bool
	Baudrate uint32
}

func ParseDefinition(path string) (*Definition, error) {
	data, error := os.ReadFile(path)

	if error != nil {
		return nil, fmt.Errorf("read definition: %w", error)
	}

	return ParseDefinitionString(string(data));
}

func ParseDefinitionString(data string) (*Definition, error) {
	definition := &Definition{
		Custom: make(map[string]string),
	}

	section := ""

	lines := strings.Split(data, "\n")

	for lineNumber, line := range lines {
		lineNumber++
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			
		}
	}
}