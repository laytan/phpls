package visitor

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/token"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phpversion"
	"golang.org/x/exp/slices"
)

// The original name is `PhpStormStubsElementAvailable` but this is sometimes aliassed.
// TODO: We should ideally resolve the use statements as well.
var names = [][]byte{
	[]byte("PhpStormStubsElementAvailable"),
	[]byte("Available"),
	[]byte("ElementAvailable"),
}

type ElementAvailableAttribute struct {
	visitor.Null

	version *phpversion.PHPVersion
	logging bool
}

func NewElementAvailableAttribute(
	version *phpversion.PHPVersion,
	logging bool,
) *ElementAvailableAttribute {
	return &ElementAvailableAttribute{
		version: version,
		logging: logging,
	}
}

func (e *ElementAvailableAttribute) Root(n *ast.Root) {
	n.Stmts = e.filterStmts(n.Stmts)
}

func (e *ElementAvailableAttribute) StmtNamespace(n *ast.StmtNamespace) {
	n.Stmts = e.filterStmts(n.Stmts)
}

func (e *ElementAvailableAttribute) StmtClass(n *ast.StmtClass) {
	n.Stmts = e.filterStmts(n.Stmts)
}

func (e *ElementAvailableAttribute) StmtInterface(n *ast.StmtInterface) {
	n.Stmts = e.filterStmts(n.Stmts)
}

func (e *ElementAvailableAttribute) StmtTrait(n *ast.StmtTrait) {
	n.Stmts = e.filterStmts(n.Stmts)
}

func (e *ElementAvailableAttribute) filterStmts(nodes []ast.Vertex) []ast.Vertex {
	newStmts := make([]ast.Vertex, 0, len(nodes))
	for _, stmt := range nodes {
		ok := true
		switch typedStmt := stmt.(type) {
		case *ast.StmtNamespace, *ast.StmtClass, *ast.StmtTrait, *ast.StmtInterface:
			stmt.Accept(e)

		case *ast.StmtFunction:
			ok = !e.shouldRemove(typedStmt.AttrGroups)
			if ok {
				params, removedParams := e.filterParams(typedStmt.Params)
				typedStmt.Params = params

				if len(removedParams) > 0 {
					e.removeParamsDocFromFunction(typedStmt, removedParams)

					// Setting this to be empty seems to make the printer
					// add/recalculate where separators go.
					typedStmt.SeparatorTkns = []*token.Token{}
				}
			}

		case *ast.StmtClassMethod:
			ok = !e.shouldRemove(typedStmt.AttrGroups)
			if ok {
				params, removedParams := e.filterParams(typedStmt.Params)
				typedStmt.Params = params

				if len(removedParams) > 0 {
					e.removeParamsDocFromMethod(typedStmt, removedParams)

					// Setting this to be empty seems to make the printer
					// add/recalculate where separators go.
					typedStmt.SeparatorTkns = []*token.Token{}
				}
			}

		case *ast.StmtPropertyList:
			ok = !e.shouldRemove(typedStmt.AttrGroups)
		}

		if ok {
			newStmts = append(newStmts, stmt)
			continue
		}

		e.logRemoval(stmt)
	}

	return newStmts
}

func (e *ElementAvailableAttribute) filterParams(
	params []ast.Vertex,
) (newParams []ast.Vertex, removedParams []string) {
	// PERF: should only create a new slice if we actually are removing a parameter.
	newParams = make([]ast.Vertex, 0, len(params))
	for _, param := range params {
		if typedParam, ok := param.(*ast.Parameter); ok {
			if !e.shouldRemove(typedParam.AttrGroups) {
				newParams = append(newParams, param)
				continue
			}

			paramName := string(typedParam.Var.(*ast.ExprVariable).Name.(*ast.Identifier).Value)
			// TODO: don't know what is going on here.
			// if typedParam.VariadicTkn != nil {
			// 	paramName = "..." + paramName
			// }

			removedParams = append(removedParams, paramName)

			e.logRemoval(param)
		}
	}

	return newParams, removedParams
}

func (e *ElementAvailableAttribute) shouldRemove(attrGroups []ast.Vertex) bool {
	for _, attrGroup := range attrGroups {
	Attributes:
		for _, attr := range attrGroup.(*ast.AttributeGroup).Attrs {
			if len(attr.(*ast.Attribute).Args) == 0 {
				continue
			}

			attrName := attr.(*ast.Attribute).Name.(*ast.Name)
			if len(attrName.Parts) != 1 {
				continue
			}

			if attrNamePart, ok := attrName.Parts[0].(*ast.NamePart); ok {
				var match bool
				for _, name := range names {
					if bytes.Equal(attrNamePart.Value, name) {
						match = true
						break
					}
				}

				if !match {
					continue
				}
			}

			for i, arg := range attr.(*ast.Attribute).Args {
				if i > 1 {
					break
				}

				var n []byte
				var v *phpversion.PHPVersion
				if argName, ok := arg.(*ast.Argument).Name.(*ast.Identifier); ok {
					n = argName.Value
				}

				if exprStr, ok := arg.(*ast.Argument).Expr.(*ast.ScalarString); ok {
					versionStr := strings.Trim(string(exprStr.Value), `'"`)
					if version, ok := phpversion.FromString(versionStr); ok {
						v = version
					}
				}

				if v == nil {
					continue Attributes
				}

				if bytes.Equal(n, []byte("from")) || i == 0 {
					if v.IsHigherThan(e.version) {
						return true
					}
				}

				if bytes.Equal(n, []byte("to")) || i == 1 {
					if e.version.IsHigherThan(v) {
						return true
					}
				}
			}
		}
	}

	return false
}

func (e *ElementAvailableAttribute) removeParamsDocFromFunction(
	function *ast.StmtFunction,
	params []string,
) {
	function.FunctionTkn.FreeFloating = e.removeParamsDoc(function.FunctionTkn.FreeFloating, params)

	for _, attrGroup := range function.AttrGroups {
		attrGroup.(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating = e.removeParamsDoc(
			attrGroup.(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating, params,
		)
	}
}

func (e *ElementAvailableAttribute) removeParamsDocFromMethod(
	function *ast.StmtClassMethod,
	params []string,
) {
	function.FunctionTkn.FreeFloating = e.removeParamsDoc(function.FunctionTkn.FreeFloating, params)

	for _, attrGroup := range function.AttrGroups {
		attrGroup.(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating = e.removeParamsDoc(
			attrGroup.(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating, params,
		)
	}
}

func (e *ElementAvailableAttribute) removeParamsDoc(
	freefloatings []*token.Token,
	params []string,
) []*token.Token {
	for _, t := range freefloatings {
		if t.ID != token.T_DOC_COMMENT && t.ID != token.T_COMMENT {
			continue
		}

		nodes, err := phpdoxer.ParseDoc(string(t.Value))
		if err != nil {
			log.Println(err)
		}

		newNodes := make([]phpdoxer.Node, 0, len(nodes))
		for _, node := range nodes {
			if paramNode, ok := node.(*phpdoxer.NodeParam); ok {
				if slices.Contains(params, paramNode.Name) {
					e.logRemoval(nil)
					continue
				}
			}

			newNodes = append(newNodes, node)
		}

		// TODO: also have non-nodes/raw comments reconstructed.
		inner := strings.Join(
			functional.Map(newNodes, func(node phpdoxer.Node) string {
				// TODO: make node.String() return the raw value that was initially parsed.
				return " * " + node.String()
			}),
			"\n",
		)
		t.Value = []byte("/**\n" + inner + "\n */")
	}

	return freefloatings
}

// TODO: don't require arg.
func (e *ElementAvailableAttribute) logRemoval(n ast.Vertex) {
	if e.logging {
		_, _ = fmt.Printf("x") //nolint:forbidigo // For visualization.
	}
}
