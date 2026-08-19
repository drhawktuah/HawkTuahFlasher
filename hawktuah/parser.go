package hawktuah

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	ValueString ValueType = iota
	ValueNumber
	ValueBoolean
)

type Parser struct {
	tokens 				 []Token
	index  				   int

	pendingDocumentation []DocumentationComment
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func Parse(source string) (*Definition, error) {
	lexer := NewLexer(source)

	tokens, err := lexer.Lex()
	if err != nil {
		return nil, err
	}

	parser := NewParser(tokens)
	return parser.Parse()
}

func (parser *Parser) Parse() (*Definition, error) {
	definition := NewDefinition()

	for !parser.atEnd() {
		parser.consumeComments()

		if parser.atEnd() {
			break
		}

		token := parser.current()

		if token.Type != TokenIdentifier {
			return nil, parser.errorf(token, "expected property or section, got %s", token.Type)
		}

		if parser.peekType() == TokenColon {
			if err := parser.parseSection(definition); err != nil {
				return nil, err
			}

			continue
		}

		if err := parser.parseRootProperty(definition); err != nil {
			return nil, err
		}
	}

	if err := ValidateDefinition(definition); err != nil {
		return nil, err
	}

	return definition, nil
}

func ValidateDefinition(definition *Definition) error {
	if definition == nil {
		return fmt.Errorf("definition is nil")
	}

	if definition.Name == "" {
		return fmt.Errorf("missing required property: name")
	}

	if definition.Vendor == "" {
		return fmt.Errorf("missing required property: vendor")
	}

	if definition.Family == "" {
		return fmt.Errorf("missing required property: family")
	}

	if len(definition.Detect.VIDs) == 0 {
		return fmt.Errorf("no USB VID detection entries defined")
	}

	for index, detection := range definition.Detect.VIDs {
		if detection.Value == 0 {
			return fmt.Errorf("detect.vid[%d] cannot be 0x0000", index)
		}
	}

	if definition.Protocol.Bootloader == "" {
		return fmt.Errorf("missing required property: protocol.bootloader")
	}

	if definition.Flash.Baudrate == 0 {
		return fmt.Errorf("missing required property: flash.baudrate")
	}

	return nil
}

func (parser *Parser) parseRootProperty(definition *Definition) error {
	key := parser.current()

	parser.advance()

	if err := parser.expect(TokenEquals); err != nil {
		return err
	}

	value := parser.current()
	parser.advance()

	switch key.Value {
		case "name":
			if value.Type != TokenString {
				return parser.expected(value, TokenString)
			}

			definition.Name = value.Value
			parser.applyPendingDocumentation(&definition.NameDocumentation)

		case "vendor":
			if value.Type != TokenString {
				return parser.expected(value, TokenString)
			}

			definition.Vendor = value.Value
			parser.applyPendingDocumentation(&definition.VendorDocumentation)

		case "family":
			if value.Type != TokenString {
				return parser.expected(value, TokenString)
			}

			definition.Family = value.Value
			parser.applyPendingDocumentation(&definition.FamilyDocumentation)

		default:
			return parser.errorf(key, "unknown root property %q", key.Value)
	}

	return nil
}

func (parser *Parser) parseSection(definition *Definition) error {
	section := parser.current()
	parser.advance()

	if err := parser.expect(TokenColon); err != nil {
		return err
	}

	switch strings.ToLower(section.Value) {
		case "detect":
			return parser.parseDetect(definition)

		case "protocol":
			return parser.parseProtocol(definition)

		case "flash":
			return parser.parseFlash(definition)

		case "custom":
			return parser.parseCustom(definition)

		case "cmake":
			return parser.parseCMake(definition)

		default:
			return parser.errorf(section, "unknown section %q", section.Value)
	}
}

func (parser *Parser) parseDetect(definition *Definition) error {
	for !parser.atEnd() {
		parser.consumeComments()

		if parser.atEnd() {
			return nil
		}

		if parser.isNextSection() {
			return nil
		}

		key := parser.current()

		if key.Type != TokenIdentifier {
			return parser.errorf(key, "expected detection property, got %s", key.Type)
		}

		parser.advance()

		if err := parser.expect(TokenEquals); err != nil {
			return err
		}

		value := parser.current()
		parser.advance()

		switch strings.ToLower(key.Value) {
			case "vid":
				if value.Type != TokenNumber {
					return parser.expected(value, TokenNumber)
				}

				vid, err := parseUint16(value.Value)
				if err != nil {
					return parser.errorf(value, "invalid VID %q: %v", value.Value, err)
				}

				detection := DetectionVID{
					Value:         vid,
					Documentation: parser.takePendingDocumentation(),
				}

				definition.Detect.VIDs = append(definition.Detect.VIDs, detection)

			default:
				return parser.errorf(key, "unknown detect property %q", key.Value)
		}
	}

	return nil
}

func (parser *Parser) parseProtocol(definition *Definition) error {
	for !parser.atEnd() {
		parser.consumeComments()

		if parser.atEnd() {
			return nil
		}

		if parser.isNextSection() {
			return nil
		}

		key := parser.current()

		if key.Type != TokenIdentifier {
			return parser.errorf(key, "expected protocol property, got %s", key.Type)
		}

		parser.advance()

		if err := parser.expect(TokenEquals); err != nil {
			return err
		}

		value := parser.current()
		parser.advance()

		switch strings.ToLower(key.Value) {
			case "bootloader":
				if value.Type != TokenString {
					return parser.expected(value, TokenString)
				}

				definition.Protocol.Bootloader = value.Value
				definition.Protocol.BootloaderDocumentation = parser.takePendingDocumentation()

			default:
				return parser.errorf(key, "unknown protocol property %q", key.Value)
		}
	}

	return nil
}

func (parser *Parser) parseFlash(definition *Definition) error {
	for !parser.atEnd() {
		parser.consumeComments()

		if parser.atEnd() {
			return nil
		}

		if parser.isNextSection() {
			return nil
		}

		key := parser.current()

		if key.Type != TokenIdentifier {
			return parser.errorf(key, "expected flash property, got %s", key.Type)
		}

		parser.advance()

		if err := parser.expect(TokenEquals); err != nil {
			return err
		}

		value := parser.current()
		parser.advance()

		switch strings.ToLower(key.Value) {
			case "baudrate":
				if value.Type != TokenNumber {
					return parser.expected(value, TokenNumber)
				}

				baudrate, err := parseUint32(value.Value)
				if err != nil {
					return parser.errorf(value, "invalid baudrate %q: %v", value.Value, err)
				}

				definition.Flash.Baudrate = baudrate
				definition.Flash.BaudrateDocumentation = parser.takePendingDocumentation()

			case "erase":
				if value.Type != TokenBoolean {
					return parser.expected(value, TokenBoolean)
				}

				erase, err := strconv.ParseBool(value.Value)
				if err != nil {
					return parser.errorf(value, "invalid erase value %q: %v", value.Value, err)
				}

				definition.Flash.Erase = erase
				definition.Flash.EraseDocumentation = parser.takePendingDocumentation()

			case "verify":
				if value.Type != TokenBoolean {
					return parser.expected(value, TokenBoolean)
				}

				verify, err := strconv.ParseBool(value.Value)
				if err != nil {
					return parser.errorf(value, "invalid verify value %q: %v", value.Value, err)
				}

				definition.Flash.Verify = verify
				definition.Flash.VerifyDocumentation = parser.takePendingDocumentation()

			default:
				return parser.errorf(key, "unknown flash property %q", key.Value)
		}
	}

	return nil
}

func (parser *Parser) parseCustom(definition *Definition) error {
	return parser.parseGenericProperties(definition.Custom, "custom")
}

func (parser *Parser) parseCMake(definition *Definition) error {
	return parser.parseGenericProperties(definition.CMake, "cmake")
}

func (parser *Parser) parseGenericProperties(properties map[string]Value, section string) error {
	for !parser.atEnd() {
		parser.consumeComments()

		if parser.atEnd() {
			return nil
		}

		if parser.isNextSection() {
			return nil
		}

		key := parser.current()

		if key.Type != TokenIdentifier {
			return parser.errorf(key, "expected property in %s section, got %s", section, key.Type)
		}

		parser.advance()

		if err := parser.expect(TokenEquals); err != nil {
			return err
		}

		token := parser.current()

		value, err := parser.parseGenericValue(token)
		if err != nil {
			return err
		}

		parser.advance()

		value.Documentation = parser.takePendingDocumentation()

		if _, exists := properties[key.Value]; exists {
			return parser.errorf(key, "duplicate property %q in %s section", key.Value, section)
		}

		properties[key.Value] = value
	}

	return nil
}

func (parser *Parser) parseGenericValue(token Token) (Value, error) {
	switch token.Type {
		case TokenString:
			return Value{
				Type:   ValueString,
				String: token.Value,
			}, nil

		case TokenNumber:
			number, err := strconv.ParseUint(token.Value, 0, 64)

			if err != nil {
				return Value{}, parser.errorf(token, "invalid numeric value %q: %v", token.Value, err)
			}

			return Value{
				Type:   ValueNumber,
				Number: number,
			}, nil

		case TokenBoolean:
			boolean, err := strconv.ParseBool(token.Value)

			if err != nil {
				return Value{}, parser.errorf(token, "invalid boolean value %q: %v", token.Value, err)
			}

			return Value{
				Type:    ValueBoolean,
				Boolean: boolean,
			}, nil

		default:
			return Value{}, parser.expected(token, TokenString)
	}
}

func (parser *Parser) consumeComments() {
	for !parser.atEnd() {
		switch parser.current().Type {
		case TokenComment:
			parser.advance()

		case TokenDocumentation:
			if parser.current().Documentation != nil {
				parser.pendingDocumentation = append(parser.pendingDocumentation, *parser.current().Documentation)
			}

			parser.advance()

		default:
			return
		}
	}
}

func (parser *Parser) applyPendingDocumentation(documentation *Documentation) {
	if documentation.Tags == nil {
		documentation.Tags = make(map[string]string)
	}

	for _, comment := range parser.pendingDocumentation {
		documentation.Tags[comment.Tag] = comment.Value
	}

	parser.pendingDocumentation = nil
}

func (parser *Parser) takePendingDocumentation() PropertyDocumentation {
	documentation := NewPropertyDocumentation()

	for _, comment := range parser.pendingDocumentation {
		documentation.Tags[comment.Tag] = comment.Value
	}

	parser.pendingDocumentation = nil

	return documentation
}

func (parser *Parser) isNextSection() bool {
	if parser.current().Type != TokenIdentifier {
		return false
	}

	return parser.peekType() == TokenColon
}

func (parser *Parser) expect(expected TokenType) error {
	token := parser.current()

	if token.Type != expected {
		return parser.errorf(token, "expected %s, got %s", expected, token.Type)
	}

	parser.advance()

	return nil
}

func (parser *Parser) expected(token Token, expected TokenType) error {
	return parser.errorf(token, "expected %s, got %s", expected, token.Type)
}

func (parser *Parser) current() Token {
	if parser.index >= len(parser.tokens) {
		return Token{
			Type: TokenEOF,
		}
	}

	return parser.tokens[parser.index]
}

func (parser *Parser) peek() Token {
	if parser.index + 1 >= len(parser.tokens) {
		return Token{
			Type: TokenEOF,
		}
	}

	return parser.tokens[parser.index + 1]
}

func (parser *Parser) peekType() TokenType {
	return parser.peek().Type
}

func (parser *Parser) advance() {
	if parser.index < len(parser.tokens) {
		parser.index++
	}
}

func (parser *Parser) atEnd() bool {
	return parser.current().Type == TokenEOF
}

func (parser *Parser) errorf(token Token, format string, arguments ...any) error {
	return fmt.Errorf("%d:%d: %s", token.Position.Line, token.Position.Column, fmt.Sprintf(format, arguments...))
}

func parseUint16(value string) (uint16, error) {
	number, err := strconv.ParseUint(strings.TrimSpace(value), 0, 16)

	if err != nil {
		return 0, err
	}

	return uint16(number), nil
}

func parseUint32(value string) (uint32, error) {
	number, err := strconv.ParseUint(strings.TrimSpace(value), 0, 32)

	if err != nil {
		return 0, err
	}

	return uint32(number), nil
}
