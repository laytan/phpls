package project

import (
	"errors"
	"fmt"
	"time"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/noverify/src/ir/irconv"
	"github.com/VKCOM/php-parser/pkg/conf"
	"github.com/VKCOM/php-parser/pkg/parser"
)

type File struct {
	content  string
	modified time.Time
	path     string
}

func (f *File) parse(config conf.Config) (*ir.Root, error) {
	rootNode, err := parser.Parse([]byte(f.content), config)
	if err != nil {
		return nil, fmt.Errorf("Error parsing file into AST: %w", err)
	}

	irNode := irconv.ConvertNode(rootNode)
	irRootNode, ok := irNode.(*ir.Root)
	if !ok {
		return nil, errors.New("AST root node could not be converted to IR root node")
	}

	return irRootNode, nil
}
