package parsing

import (
	"io/fs"

	"github.com/VKCOM/noverify/src/ir"
)

type Parser interface {
	Parse(content string) (*ir.Root, error)

	Read(path string) (string, error)

	Walk(root string, walker func(path string, d fs.DirEntry, err error) error) error
}
