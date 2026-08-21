package hawktuah_test

import (
	"flasher/hawktuah"
	"testing"
)

func TestLexerIdentifiers(t *testing.T) {
	source := `
name = "ESP32"
vendor = "Espressif"
family = "ESP32"
`

	tokens, err := hawktuah.NewLexer(source).Lex()
	if err != nil {
		t.Fatalf("Lex() failed: %v", err)
	}

	expected := [] hawktuah.TokenType{
		hawktuah.TokenIdentifier,
		hawktuah.TokenEquals,
		hawktuah.TokenString,

		hawktuah.TokenIdentifier,
		hawktuah.TokenEquals,
		hawktuah.TokenString,

		hawktuah.TokenIdentifier,
		hawktuah.TokenEquals,
		hawktuah.TokenString,

		hawktuah.TokenEOF,
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for index, token := range tokens {
		if token.Type != expected[index] {
			t.Errorf("token %d: expected %s, got %s", index, expected[index], token.Type)
		}
	}
}

func TestLexerNumbers(t *testing.T) {
	source := `
decimal = 921600
hexadecimal = 0x303A
`

	tokens, err := hawktuah.NewLexer(source).Lex()
	if err != nil {
		t.Fatalf("Lex() failed: %v", err)
	}

	var numbers []string

	for _, token := range tokens {
		if token.Type == hawktuah.TokenNumber {
			numbers = append(numbers, token.Value)
		}
	}

	if len(numbers) != 2 {
		t.Fatalf("expected 2 numbers, got %d", len(numbers))
	}

	if numbers[0] != "921600" {
		t.Errorf("expected decimal value 921600, got %q", numbers[0])
	}

	if numbers[1] != "0x303A" {
		t.Errorf("expected hexadecimal value 0x303A, got %q", numbers[1])
	}
}

func TestLexerBooleans(t *testing.T) {
	source := `
erase = true
verify = false
`

	tokens, err := hawktuah.NewLexer(source).Lex()
	if err != nil {
		t.Fatalf("Lex() failed: %v", err)
	}

	var values []string

	for _, token := range tokens {
		if token.Type == hawktuah.TokenBoolean {
			values = append(values, token.Value)
		}
	}

	if len(values) != 2 {
		t.Fatalf("expected 2 boolean tokens, got %d", len(values))
	}

	if values[0] != "true" {
		t.Errorf("expected true, got %q", values[0])
	}

	if values[1] != "false" {
		t.Errorf("expected false, got %q", values[1])
	}
}

func TestLexerStringEscapes(t *testing.T) {
	source := `value = "hello\nworld\t\"test\""`

	tokens, err := hawktuah.NewLexer(source).Lex()
	if err != nil {
		t.Fatalf("Lex() failed: %v", err)
	}

	var value string

	for _, token := range tokens {
		if token.Type == hawktuah.TokenString {
			value = token.Value
			break
		}
	}

	expected := "hello\nworld\t\"test\""

	if value != expected {
		t.Errorf("expected %q, got %q", expected, value)
	}
}

func TestLexerComments(t *testing.T) {
	source := `
# normal comment
name = "ESP32"
`

	tokens, err := hawktuah.NewLexer(source).Lex()
	if err != nil {
		t.Fatalf("Lex() failed: %v", err)
	}

	found := false

	for _, token := range tokens {
		if token.Type == hawktuah.TokenComment {
			found = true

			if token.Value != "normal comment" {
				t.Errorf("unexpected comment value %q", token.Value)
			}

			break
		}
	}

	if !found {
		t.Fatal("expected comment token")
	}
}

func TestLexerDocumentation(t *testing.T) {
	source := `
# @description ESP32 development board
name = "ESP32"
`

	tokens, err := hawktuah.NewLexer(source).Lex()
	if err != nil {
		t.Fatalf("Lex() failed: %v", err)
	}

	var documentation *hawktuah.DocumentationComment

	for _, token := range tokens {
		if token.Type == hawktuah.TokenDocumentation {
			documentation = token.Documentation
			break
		}
	}

	if documentation == nil {
		t.Fatal("expected documentation token")
	}

	if documentation.Tag != "description" {
		t.Errorf(
			"expected documentation tag %q, got %q",
			"description",
			documentation.Tag,
		)
	}

	if documentation.Value != "ESP32 development board" {
		t.Errorf(
			"expected documentation value %q, got %q",
			"ESP32 development board",
			documentation.Value,
		)
	}
}

func TestLexerPositions(t *testing.T) {
	source := "name = \"ESP32\""

	tokens, err := hawktuah.NewLexer(source).Lex()
	if err != nil {
		t.Fatalf("Lex() failed: %v", err)
	}

	if tokens[0].Position.Line != 1 {
		t.Errorf("expected line 1, got %d", tokens[0].Position.Line)
	}

	if tokens[0].Position.Column != 1 {
		t.Errorf("expected column 1, got %d", tokens[0].Position.Column)
	}
}

func TestLexerErrors(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "unexpected character",
			source: `name = $`,
		},
		{
			name:   "unterminated string",
			source: `name = "ESP32`,
		},
		{
			name:   "unterminated escape",
			source: `name = "ESP32\`,
		},
		{
			name:   "invalid escape",
			source: `name = "\q"`,
		},
		{
			name:   "invalid hexadecimal",
			source: `vid = 0x`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := hawktuah.NewLexer(test.source).Lex()

			if err == nil {
				t.Fatal("expected lexer error")
			}
		})
	}
}