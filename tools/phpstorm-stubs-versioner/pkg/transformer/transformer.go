package transformer

import "github.com/VKCOM/php-parser/pkg/ast"

type Transformer interface {
	Transform(ast ast.Vertex)
}
