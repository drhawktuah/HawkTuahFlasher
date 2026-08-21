package dynamic

import (
	"flasher/core"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type CommandDetails struct {
	Name        string
	Description string
	Category    string
	Package     string
	Path        string
}

func DiscoverCommands(root string) ([]CommandDetails, error) {
	if root == "" {
		return nil, core.Errorf("command directory is empty")
	}

	info, err := os.Stat(root)
	if err != nil {
		return nil, core.Errorf("stat command directory: %w", err)
	}

	if !info.IsDir() {
		return nil, core.Errorf("command path %q is not a directory", root)
	}

	commands := make([]CommandDetails, 0)

	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkError error) error {
		if walkError != nil {
			return walkError
		}

		if entry.IsDir() {
			if path != root && strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}

			return nil
		}

		name := entry.Name()
		if name == "" || strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}

		details, err := parseCommandFile(root, path)
		if err != nil {
			return err
		}

		commands = append(commands, details...)

		return nil
	})

	if err != nil {
		return nil, core.Errorf("discover commands: %w", err)
	}

	sort.Slice(commands, func(i, j int) bool {
		if commands[i].Category != commands[j].Category {
			return commands[i].Category < commands[j].Category
		}

		if commands[i].Name != commands[j].Name {
			return commands[i].Name < commands[j].Name
		}

		return commands[i].Path < commands[j].Path
	})

	return commands, nil
}

func parseCommandFile(root string, path string) ([]CommandDetails, error) {
	fileSet := token.NewFileSet()

	file, err := goparser.ParseFile(fileSet, path, nil, goparser.ParseComments)
	if err != nil {
		return nil, core.Errorf("parse command file %q: %w", path, err)
	}

	relativePath, err := filepath.Rel(root, path)
	if err != nil {
		return nil, core.Errorf("resolve command path %q: %w", path, err)
	}

	category := filepath.Dir(relativePath)

	if category == "." {
		category = ""
	}

	category = filepath.ToSlash(category)

	commands := make([]CommandDetails, 0)

	ast.Inspect(file, func(node ast.Node) bool {
		composite, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}

		if !isCommandComposite(composite) {
			return true
		}

		details := CommandDetails{
			Category: category,
			Package:  file.Name.Name,
			Path:     path,
		}

		for _, element := range composite.Elts {
			field, ok := element.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			key, ok := field.Key.(*ast.Ident)
			if !ok {
				continue
			}

			switch key.Name {
				case "Name":
					if value, ok := stringValue(field.Value); ok {
						details.Name = value
					}

				case "Description":
					if value, ok := stringValue(field.Value); ok {
						details.Description = value
					}
			}
		}

		if details.Name != "" {
			commands = append(commands, details)
		}

		return true
	})

	return commands, nil
}

func isCommandComposite(composite *ast.CompositeLit) bool {
	switch expression := composite.Type.(type) {
		case *ast.Ident:
			return expression.Name == "Command"

		case *ast.SelectorExpr:
			packageIdentifier, ok := expression.X.(*ast.Ident)
			if !ok {
				return false
			}

			return expression.Sel.Name == "Command" &&
				packageIdentifier.Name == "commandline"

	default:
		return false
	}
}

func stringValue(expression ast.Expr) (string, bool) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok {
		return "", false
	}

	if literal.Kind != token.STRING {
		return "", false
	}

	value, err := unquoteString(literal.Value)
	if err != nil {
		return "", false
	}

	return value, true
}

func unquoteString(value string) (string, error) {
	if len(value) < 2 {
		return "", core.Errorf("invalid string literal")
	}

	switch value[0] {
		case '"':
			return strconv.Unquote(value)

		case '`':
			return value[1 : len(value)-1], nil

		default:
			return "", core.Errorf("unsupported string literal")
	}
}

func lowerFirst(value string) string {
	if value == "" {
		return value
	}

	return strings.ToLower(value[:1]) + value[1:]
}