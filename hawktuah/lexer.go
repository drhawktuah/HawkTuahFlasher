package hawktuah

import (
	"flasher/core"
	"strings"
	"unicode"
)

type TokenType uint8

const (
	CommentPrefix       = '#'
	DocumentationPrefix = '@'
)

const (
	TokenUnknown TokenType = iota

	TokenIdentifier
	TokenString
	TokenNumber
	TokenBoolean

	TokenEquals
	TokenColon

	TokenComment
	TokenDocumentation

	TokenEOF
)

type Position struct {
	Offset int
	Line   int
	Column int
}

type DocumentationComment struct {
	Tag   string
	Value string

	Position Position
}

type Token struct {
	Type  TokenType
	Value string

	Position Position

	Documentation *DocumentationComment
}

type Lexer struct {
	source []rune

	offset int
	line   int
	column int
}

func (tokenType TokenType) String() string {
	switch tokenType {
		case TokenUnknown:
			return "unknown"

		case TokenIdentifier:
			return "identifier"

		case TokenString:
			return "string"

		case TokenNumber:
			return "number"

		case TokenBoolean:
			return "boolean"

		case TokenEquals:
			return "="

		case TokenColon:
			return ":"

		case TokenComment:
			return "comment"

		case TokenDocumentation:
			return "documentation"

		case TokenEOF:
			return "end-of-file"

		default:
			return "unknown"
	}
}

func NewLexer(source string) *Lexer {
	return &Lexer{
		source: []rune(source),
		line:   1,
		column: 1,
	}
}

func (lexer *Lexer) Lex() ([]Token, error) {
	tokens := make([]Token, 0)

	for {
		token, err := lexer.Next()
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)

		if token.Type == TokenEOF {
			break
		}
	}

	return tokens, nil
}

func (lexer *Lexer) Next() (Token, error) {
	lexer.skipWhitespace()

	if lexer.atEnd() {
		return Token{
			Type:     TokenEOF,
			Position: lexer.position(),
		}, nil
	}

	position := lexer.position()
	character := lexer.current()

	switch character {
		case CommentPrefix:
			return lexer.lexComment()

		case '=':
			lexer.advance()

			return Token{
				Type:     TokenEquals,
				Value:    "=",
				Position: position,
			}, nil

		case ':':
			lexer.advance()

			return Token{
				Type:     TokenColon,
				Value:    ":",
				Position: position,
			}, nil

		case '"':
			return lexer.lexString()

		default:
			switch {
				case isIdentifierStart(character):
					return lexer.lexIdentifier()

				case isNumberStart(character):
					return lexer.lexNumber()

				default:
					return Token{}, lexer.errorf(position, "unexpected character %q", character)
				}
	}
}

func (lexer *Lexer) lexComment() (Token, error) {
	position := lexer.position()

	lexer.advance()

	for !lexer.atEnd() {
		character := lexer.current()

		if character == ' ' || character == '\t' {
			lexer.advance()
			continue
		}

		break
	}

	start := lexer.offset

	for !lexer.atEnd() {
		character := lexer.current()

		if character == '\n' {
			break
		}

		lexer.advance()
	}

	value := string(lexer.source[start:lexer.offset])
	value = strings.TrimSpace(value)

	if strings.HasPrefix(value, string(DocumentationPrefix)) {
		documentation, err := parseDocumentation(value, position)

		if err != nil {
			return Token{}, err
		}

		return Token{
			Type:          TokenDocumentation,
			Value:         value,
			Position:      position,
			Documentation: &documentation,
		}, nil
	}

	return Token{
		Type:     TokenComment,
		Value:    value,
		Position: position,
	}, nil
}

func parseDocumentation(value string, position Position) (DocumentationComment, error) {
	value = strings.TrimSpace(value)

	if !strings.HasPrefix(value, string(DocumentationPrefix)) {
		return DocumentationComment{}, core.Errorf("invalid documentation comment")
	}

	value = strings.TrimPrefix(value, string(DocumentationPrefix))
	value = strings.TrimSpace(value)

	if value == "" {
		return DocumentationComment{}, core.Errorf("documentation tag cannot be empty at %d:%d", position.Line, position.Column)
	}

	tagEnd := strings.IndexAny(value, " \t")

	if tagEnd < 0 {
		return DocumentationComment{
			Tag:      value,
			Position: position,
		}, nil
	}

	tag := strings.TrimSpace(value[:tagEnd])
	content := strings.TrimSpace(value[tagEnd:])

	if tag == "" {
		return DocumentationComment{}, core.Errorf("documentation tag cannot be empty at %d:%d", position.Line, position.Column)
	}

	return DocumentationComment{
		Tag:      tag,
		Value:    content,
		Position: position,
	}, nil
}

func (lexer *Lexer) lexIdentifier() (Token, error) {
	position := lexer.position()
	start := lexer.offset

	for !lexer.atEnd() {
		character := lexer.current()

		if !isIdentifierCharacter(character) {
			break
		}

		lexer.advance()
	}

	value := string(lexer.source[start:lexer.offset])

	tokenType := TokenIdentifier

	switch value {
		case "true", "false":
			tokenType = TokenBoolean
	}

	return Token{
		Type:     tokenType,
		Value:    value,
		Position: position,
	}, nil
}

func (lexer *Lexer) lexNumber() (Token, error) {
	position := lexer.position()

	start := lexer.offset

	if lexer.current() == '0' && lexer.peek() != 0 && (lexer.peek() == 'x' || lexer.peek() == 'X') {
		lexer.advance()
		lexer.advance()

		digitStart := lexer.offset

		for !lexer.atEnd() {
			if !isHexDigit(lexer.current()) {
				break
			}

			lexer.advance()
		}

		if lexer.offset == digitStart {
			return Token{}, lexer.errorf(position, "expected hexadecimal digits")
		}

		return Token{
			Type:     TokenNumber,
			Value:    string(lexer.source[start:lexer.offset]),
			Position: position,
		}, nil
	}

	for !lexer.atEnd() {
		character := lexer.current()

		if character < '0' || character > '9' {
			break
		}

		lexer.advance()
	}

	return Token{
		Type:     TokenNumber,
		Value:    string(lexer.source[start:lexer.offset]),
		Position: position,
	}, nil
}

func (lexer *Lexer) lexString() (Token, error) {
	position := lexer.position()

	lexer.advance()

	var builder strings.Builder

	for !lexer.atEnd() {
		character := lexer.current()

		if character == '"' {
			lexer.advance()

			return Token{
				Type:     TokenString,
				Value:    builder.String(),
				Position: position,
			}, nil
		}

		if character == '\\' {
			lexer.advance()

			if lexer.atEnd() {
				return Token{}, lexer.errorf(position, "unterminated escape sequence")
			}

			escaped := lexer.current()

			switch escaped {
				case '\\':
					builder.WriteRune('\\')

				case '"':
					builder.WriteRune('"')

				case 'n':
					builder.WriteRune('\n')

				case 'r':
					builder.WriteRune('\r')

				case 't':
					builder.WriteRune('\t')

				default:
					return Token{}, lexer.errorf(lexer.position(), "invalid escape sequence \\%c", escaped)
			}

			lexer.advance()
			continue
		}

		if character == '\n' {
			return Token{}, lexer.errorf(position, "unterminated string")
		}

		builder.WriteRune(character)
		lexer.advance()
	}

	return Token{}, lexer.errorf(position, "unterminated string")
}

func (lexer *Lexer) skipWhitespace() {
	for !lexer.atEnd() {
		character := lexer.current()

		if unicode.IsSpace(character) {
			lexer.advance()
			continue
		}

		break
	}
}

func (lexer *Lexer) current() rune {
	if lexer.atEnd() {
		return 0
	}

	return lexer.source[lexer.offset]
}

func (lexer *Lexer) peek() rune {
	if lexer.offset+1 >= len(lexer.source) {
		return 0
	}

	return lexer.source[lexer.offset+1]
}

func (lexer *Lexer) advance() {
	if lexer.atEnd() {
		return
	}

	character := lexer.source[lexer.offset]

	lexer.offset++

	if character == '\n' {
		lexer.line++
		lexer.column = 1
	} else {
		lexer.column++
	}
}

func (lexer *Lexer) atEnd() bool {
	return lexer.offset >= len(lexer.source)
}

func (lexer *Lexer) position() Position {
	return Position{
		Offset: lexer.offset,
		Line:   lexer.line,
		Column: lexer.column,
	}
}

func (lexer *Lexer) errorf(position Position, format string, arguments ...any) error {
	return core.Errorf("%d:%d: %s", position.Line, position.Column, core.Sprintf(format, arguments...))
}

func isIdentifierStart(character rune) bool {
	return character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character == '_'
}

func isIdentifierCharacter(character rune) bool {
	return isIdentifierStart(character) || character >= '0' && character <= '9' || character == '-' || character == '.'
}

func isNumberStart(character rune) bool {
	return character >= '0' && character <= '9'
}

func isHexDigit(character rune) bool {
	return character >= '0' && character <= '9' || character >= 'a' && character <= 'f' || character >= 'A' && character <= 'F'
}
