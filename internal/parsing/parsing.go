package parsing

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/errors"
	astParser "github.com/VKCOM/php-parser/pkg/parser"
	"github.com/VKCOM/php-parser/pkg/version"
	"github.com/laytan/elephp/pkg/phpversion"
)

const (
	ErrReadFmt  = "Could not read file at %s: %w"
	ErrASTFmt   = "Error parsing file into AST: %w"
	ErrWalkFmt  = "Error arose during walk of root %s: %w"
	ErrParseFmt = "Parse/syntax error during parse: %s"
)

// Parser is responsible for and a central place for
// file system access and parsing file content into
// IR for use everywhere else.
type Parser interface {
	Parse(content string) (*ir.Root, error)

	Read(path string) (string, error)

	Walk(root string, walker func(path string, d fs.DirEntry, err error) error) error
}

type parser struct {
	config conf.Config
}

func New(phpv phpversion.PHPVersion) Parser {
	return &parser{
		config: conf.Config{
			Version: &version.Version{
				Major: uint64(phpv.Major),
				Minor: uint64(phpv.Minor),
			},
			// TODO: return these errors.
			ErrorHandlerFunc: func(e *errors.Error) {
				log.Println(fmt.Errorf(ErrParseFmt, e.String()))
			},
		},
	}
}

func (p *parser) Parse(content string) (*ir.Root, error) {
	ast, err := astParser.Parse([]byte(content), p.config)
	if err != nil {
		return nil, fmt.Errorf(ErrASTFmt, err)
	}

	irNode := irconv.ConvertNode(ast)
	root, ok := irNode.(*ir.Root)
	if !ok {
		panic("Top level node is not the root")
	}

	return root, nil
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
		return fmt.Errorf(ErrWalkFmt, root, err)
	}

	return nil
}
