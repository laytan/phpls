package parsing

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"appliedgo.net/what"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/conf"
	astErrors "github.com/laytan/php-parser/pkg/errors"
	"github.com/laytan/php-parser/pkg/lexer"
	astParser "github.com/laytan/php-parser/pkg/parser"
	"github.com/laytan/php-parser/pkg/version"
)

const (
	ErrReadFmt = "Could not read file at %s: %w"
	ErrASTFmt  = "Error parsing file into AST: %w"
	ErrWalkFmt = "Error arose during walk of root %s: %w"
)

var ErrNoContent = errors.New(
	"No AST could be parsed, there were unrecoverable syntax errors or the file was empty",
)

// Parser is responsible for and a central place for
// file system access and parsing file content into
// IR for use everywhere else.
type Parser interface {
	Parse(content []byte) (*ast.Root, error)

	Lexer(content []byte) (lexer.Lexer, error)

	Read(path string) (string, error)

	Walk(root string, walker func(path string, d fs.DirEntry, err error) error) error
}

type parser struct {
	config conf.Config
}

func New(phpv *phpversion.PHPVersion) Parser {
	return &parser{
		config: conf.Config{
			Version: &version.Version{
				Major: uint64(phpv.Major),
				Minor: uint64(phpv.Minor),
			},
			// TODO: return these errors.
			ErrorHandlerFunc: func(e *astErrors.Error) {
				what.Happens(e.String())
			},
		},
	}
}

func (p *parser) Parse(content []byte) (*ast.Root, error) {
	a, err := astParser.Parse(content, p.config)
	if err != nil || a == nil {
		if err == nil {
			return nil, ErrNoContent
		}

		return nil, fmt.Errorf(ErrASTFmt, err)
	}

	return a.(*ast.Root), nil
}

func (p *parser) Lexer(content []byte) (lexer.Lexer, error) {
	res, err := lexer.New(content, p.config)
	if err != nil {
		return nil, fmt.Errorf("creating lexer: %w", err)
	}

	return res, nil
}

func (p *parser) Read(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf(ErrReadFmt, path, err)
	}

	return string(content), nil
}

func (p *parser) Walk(root string, walker func(path string, d fs.DirEntry, err error) error) error {
	err := filepath.WalkDir(root, walker)
	if err != nil {
		if err == filepath.SkipDir { //nolint:errorlint // stdlib checks == so don't want wrapping.
			return err //nolint:wrapcheck // stdlib checks == so don't want wrapping.
		}

		return fmt.Errorf(ErrWalkFmt, root, err)
	}

	return nil
}
