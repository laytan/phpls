package transformer

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/tools/phpstorm-stubs-versioner/pkg/visitor"
)

// ElementAvailableAttribute removes nodes with the PhpStormStubsElementAvailable
// attribute that does not match the given version.
type ElementAvailableAttribute struct {
	visitor *visitor.ElementAvailableAttribute
}

func NewElementAvailableAttribute(version *phpversion.PHPVersion, logger Logger) *ElementAvailableAttribute {
	return &ElementAvailableAttribute{
		visitor: visitor.NewElementAvailableAttribute(version, logger),
	}
}

func (r *ElementAvailableAttribute) Transform(ast ast.Vertex) {
	ast.Accept(r.visitor)
}
