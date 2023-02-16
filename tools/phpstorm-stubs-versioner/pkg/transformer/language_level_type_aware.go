package transformer

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/tools/phpstorm-stubs-versioner/pkg/visitor"
)

// LanguageLevelTypeAware applies the types in the LanguageLevelTypeAware args.
type LanguageLevelTypeAware struct {
	visitor *visitor.LanguageLevelTypeAware
}

func NewLanguageLevelTypeAware(version *phpversion.PHPVersion) *LanguageLevelTypeAware {
	return &LanguageLevelTypeAware{
		visitor: visitor.NewLanguageLevelTypeAware(version, true),
	}
}

func (r *LanguageLevelTypeAware) Transform(ast ast.Vertex) {
	ast.Accept(r.visitor)
}
