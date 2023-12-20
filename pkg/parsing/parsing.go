package parsing

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"appliedgo.net/what"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/conf"
	astErrors "github.com/laytan/php-parser/pkg/errors"
	"github.com/laytan/php-parser/pkg/lexer"
	astParser "github.com/laytan/php-parser/pkg/parser"
	"github.com/laytan/php-parser/pkg/version"
	"github.com/laytan/phpls/pkg/phpversion"
)

var (
	ErrRead      = errors.New("could not read file")
	ErrAst       = errors.New("could not parse content into AST")
	ErrWalk      = errors.New("error arose walking")
	ErrNoContent = errors.New(
		"no AST could be parsed, there were unrecoverable syntax errors or the file was empty",
	)
)

type Parser struct {
	config conf.Config
}

func New(phpv *phpversion.PHPVersion) *Parser {
	return &Parser{
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

func (p *Parser) Parse(content []byte) (root *ast.Root, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%w: %v", ErrAst, e)
		}
	}()

	a, err := astParser.Parse(content, p.config)
	if err != nil || a == nil {
		if err == nil {
			return nil, ErrNoContent
		}

		return nil, fmt.Errorf("%w: %w", ErrAst, err)
	}

	return a.(*ast.Root), nil
}

func (p *Parser) Lexer(content []byte) (lexer.Lexer, error) {
	res, err := lexer.New(content, p.config)
	if err != nil {
		return nil, fmt.Errorf("creating lexer: %w", err)
	}

	return res, nil
}

func (p *Parser) Read(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %q: %w: %w", path, ErrRead, err)
	}

	return string(content), nil
}

func (p *Parser) Walk(root string, walker func(path string, d fs.DirEntry, err error) error) error {
	err := filepath.WalkDir(root, walker)
	if err != nil {
		if err == filepath.SkipDir { //nolint:errorlint // stdlib checks == so don't want wrapping.
			return err //nolint:wrapcheck // stdlib checks == so don't want wrapping.
		}

		return fmt.Errorf("walking %q: %w: %w", root, ErrWalk, err)
	}

	return nil
}
