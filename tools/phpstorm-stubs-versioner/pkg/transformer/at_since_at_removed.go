package transformer

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/tools/phpstorm-stubs-versioner/pkg/visitor"
)

// AtSinceAtRemoved removes nodes that are
// @since > or @removed <= the given version in the PHP core stubs.
type AtSinceAtRemoved struct {
	visitor *visitor.AtSinceAtRemoved
}

func NewAtSinceAtRemoved(version *phpversion.PHPVersion) *AtSinceAtRemoved {
	return &AtSinceAtRemoved{
		visitor: visitor.NewAtSinceAtRemoved(version, true),
	}
}

func (r *AtSinceAtRemoved) Transform(ast ast.Vertex) {
	ast.Accept(r.visitor)
}
