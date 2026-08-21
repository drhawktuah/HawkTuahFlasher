package hawktuah_test

import (
	"flasher/hawktuah"
	"testing"
)

func TestParseESP32Definition(t *testing.T) {
	source := `
name = "ESP32-WROOM-32"
vendor = "Espressif"
family = "ESP32"

detect:
    vid = 0x303A

protocol:
    bootloader = "esptool"

flash:
    baudrate = 921600
    erase = true
    verify = true

custom:
    architecture = "xtensa"
    chip = "esp32"
    cores = 2
    wireless = true

cmake:
    toolchain = "xtensa-esp-elf"
    cpu_flags = "-mlongcalls -mtext-section-literals"
`

	definition, err := hawktuah.Parse(source)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if definition.Name != "ESP32-WROOM-32" {
		t.Errorf("Name = %q", definition.Name)
	}

	if definition.Vendor != "Espressif" {
		t.Errorf("Vendor = %q", definition.Vendor)
	}

	if definition.Family != "ESP32" {
		t.Errorf("Family = %q", definition.Family)
	}

	if len(definition.Detect.VIDs) != 1 {
		t.Fatalf("VID count = %d, want 1", len(definition.Detect.VIDs))
	}

	if definition.Detect.VIDs[0].Value != 0x303A {
		t.Errorf("VID = 0x%04X, want 0x303A",
			definition.Detect.VIDs[0].Value)
	}

	if definition.Protocol.Bootloader != "esptool" {
		t.Errorf("Bootloader = %q, want esptool",
			definition.Protocol.Bootloader)
	}

	if definition.Flash.Baudrate != 921600 {
		t.Errorf("Baudrate = %d, want 921600",
			definition.Flash.Baudrate)
	}

	if !definition.Flash.Erase {
		t.Error("Erase = false, want true")
	}

	if !definition.Flash.Verify {
		t.Error("Verify = false, want true")
	}

	architecture, ok := definition.Custom["architecture"]
	if !ok {
		t.Fatal("missing custom.architecture")
	}

	if architecture.Type != hawktuah.ValueString ||
		architecture.String != "xtensa" {
		t.Errorf("custom.architecture = %#v", architecture)
	}

	cores, ok := definition.Custom["cores"]
	if !ok {
		t.Fatal("missing custom.cores")
	}

	if cores.Type != hawktuah.ValueNumber ||
		cores.Number != 2 {
		t.Errorf("custom.cores = %#v", cores)
	}

	wireless, ok := definition.Custom["wireless"]
	if !ok {
		t.Fatal("missing custom.wireless")
	}

	if wireless.Type != hawktuah.ValueBoolean ||
		!wireless.Boolean {
		t.Errorf("custom.wireless = %#v", wireless)
	}
}
