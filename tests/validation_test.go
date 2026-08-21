package hawktuah_test

import (
	"flasher/hawktuah"
	"testing"
)

func validDefinition() *hawktuah.Definition {
	return &hawktuah.Definition{
		Name:   "ESP32",
		Vendor: "Espressif",
		Family: "ESP32",

		Detect: hawktuah.Detection{
			VIDs: []hawktuah.DetectionVID{
				{
					Value: 0x303A,
				},
			},
		},

		Protocol: hawktuah.Protocol{
			Bootloader: "esptool",
		},

		Flash: hawktuah.Flash{
			Baudrate: 921600,
		},
	}
}

func TestValidateDefinition(t *testing.T) {
	if err := hawktuah.ValidateDefinition(validDefinition()); err != nil {
		t.Fatalf("valid definition failed validation: %v", err)
	}
}

func TestValidateDefinitionNil(t *testing.T) {
	if err := hawktuah.ValidateDefinition(nil); err == nil {
		t.Fatal("expected nil definition error")
	}
}

func TestValidateDefinitionMissingName(t *testing.T) {
	definition := validDefinition()
	definition.Name = ""

	if err := hawktuah.ValidateDefinition(definition); err == nil {
		t.Fatal("expected missing name error")
	}
}

func TestValidateDefinitionMissingVendor(t *testing.T) {
	definition := validDefinition()
	definition.Vendor = ""

	if err := hawktuah.ValidateDefinition(definition); err == nil {
		t.Fatal("expected missing vendor error")
	}
}

func TestValidateDefinitionMissingFamily(t *testing.T) {
	definition := validDefinition()
	definition.Family = ""

	if err := hawktuah.ValidateDefinition(definition); err == nil {
		t.Fatal("expected missing family error")
	}
}

func TestValidateDefinitionMissingVID(t *testing.T) {
	definition := validDefinition()
	definition.Detect.VIDs = nil

	if err := hawktuah.ValidateDefinition(definition); err == nil {
		t.Fatal("expected missing VID error")
	}
}

func TestValidateDefinitionZeroVID(t *testing.T) {
	definition := validDefinition()
	definition.Detect.VIDs[0].Value = 0

	if err := hawktuah.ValidateDefinition(definition); err == nil {
		t.Fatal("expected zero VID error")
	}
}

func TestValidateDefinitionMissingBootloader(t *testing.T) {
	definition := validDefinition()
	definition.Protocol.Bootloader = ""

	if err := hawktuah.ValidateDefinition(definition); err == nil {
		t.Fatal("expected missing bootloader error")
	}
}

func TestValidateDefinitionMissingBaudrate(t *testing.T) {
	definition := validDefinition()
	definition.Flash.Baudrate = 0

	if err := hawktuah.ValidateDefinition(definition); err == nil {
		t.Fatal("expected missing baudrate error")
	}
}