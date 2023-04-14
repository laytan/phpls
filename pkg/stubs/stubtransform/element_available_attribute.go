package stubtransform

import (
	"bytes"
	"log"
	"strings"

	"github.com/laytan/phpls/pkg/phpdoxer"
	"github.com/laytan/phpls/pkg/phpversion"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/token"
	"github.com/laytan/php-parser/pkg/visitor"
	"golang.org/x/exp/slices"
)

// ElementAvailableAttribute removes nodes with the PhpStormStubsElementAvailable
// attribute that does not match the given version.
type ElementAvailableAttribute struct {
	visitor.Null
	version   *phpversion.PHPVersion
	targetter *targetter
	logger    Logger
}

func NewElementAvailableAttribute(
	version *phpversion.PHPVersion,
	logger Logger,
) *ElementAvailableAttribute {
	return &ElementAvailableAttribute{
		version: version,
		logger:  logger,
		targetter: newTargetter([][]byte{
			[]byte("JetBrains"),
			[]byte("PhpStorm"),
			[]byte("Internal"),
			[]byte("PhpStormStubsElementAvailable"),
		}),
	}
}

func (e *ElementAvailableAttribute) Root(n *ast.Root) {
	n.Stmts = e.filterStmts(n.Stmts)
}

func (e *ElementAvailableAttribute) StmtNamespace(n *ast.StmtNamespace) {
	exit := e.targetter.EnterNamespace(n)
	defer exit()

	n.Stmts = e.filterStmts(n.Stmts)
}

func (e *ElementAvailableAttribute) StmtUse(n *ast.StmtUseList) {
	for _, s := range n.Uses {
		s.Accept(e)
	}
}

func (e *ElementAvailableAttribute) StmtUseDeclaration(n *ast.StmtUse) {
	e.targetter.EnterUse(n)
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
		case *ast.StmtUseList, *ast.StmtNamespace, *ast.StmtClass, *ast.StmtTrait, *ast.StmtInterface:
			stmt.Accept(e)

		case *ast.StmtFunction:
			rm, newAttrGroups := e.shouldRemove(typedStmt.AttrGroups)
			ok = !rm

			if ok {
				typedStmt.AttrGroups = newAttrGroups
				params, removedParams := e.filterParams(typedStmt.Params)
				removedParamNames := e.getRemovedParamNames(typedStmt.Params, params)
				typedStmt.Params = params

				if removedParams {
					e.removeParamsDocFromFunction(typedStmt, removedParamNames)

					// Setting this to be empty seems to make the printer
					// add/recalculate where separators go.
					typedStmt.SeparatorTkns = []*token.Token{}
				}
			}

		case *ast.StmtClassMethod:
			rm, newAttrGroups := e.shouldRemove(typedStmt.AttrGroups)
			ok = !rm

			if ok {
				typedStmt.AttrGroups = newAttrGroups
				params, removedParams := e.filterParams(typedStmt.Params)
				removedParamNames := e.getRemovedParamNames(typedStmt.Params, params)
				typedStmt.Params = params

				if removedParams {
					e.removeParamsDocFromMethod(typedStmt, removedParamNames)

					// Setting this to be empty seems to make the printer
					// add/recalculate where separators go.
					typedStmt.SeparatorTkns = []*token.Token{}
				}
			}

		case *ast.StmtPropertyList:
			rm, newAttrGroups := e.shouldRemove(typedStmt.AttrGroups)
			ok = !rm

			if ok {
				typedStmt.AttrGroups = newAttrGroups
			}
		}

		if ok {
			newStmts = append(newStmts, stmt)
			continue
		}

		e.logRemoval()
	}

	return newStmts
}

func (e *ElementAvailableAttribute) filterParams(
	params []ast.Vertex,
) (newParams []ast.Vertex, removedParams bool) {
	newParams = make([]ast.Vertex, 0, len(params))
	for _, param := range params {
		if typedParam, ok := param.(*ast.Parameter); ok {
			rm, newAttrGroups := e.shouldRemove(typedParam.AttrGroups)

			if !rm {
				typedParam.AttrGroups = newAttrGroups

				newParams = append(newParams, typedParam)
				continue
			}

			e.logRemoval()
			removedParams = true
		}
	}

	return newParams, removedParams
}

// Returns all the parameters that were in 'params' but are not in 'newParams'
// to the 'removedParams' slice.
func (e *ElementAvailableAttribute) getRemovedParamNames(
	params []ast.Vertex,
	newParams []ast.Vertex,
) []string {
	removedParamNames := []string{}
	for _, param := range params {
		paramName := param.(*ast.Parameter).Var.(*ast.ExprVariable).Name.(*ast.Identifier).Value

		var inNewParams bool
		for _, newParam := range newParams {
			newParamName := newParam.(*ast.Parameter).Var.(*ast.ExprVariable).Name.(*ast.Identifier).Value
			if bytes.Equal(paramName, newParamName) {
				inNewParams = true
				break
			}
		}

		if !inNewParams {
			name := string(paramName)

			if param.(*ast.Parameter).VariadicTkn != nil {
				name = "..." + name
			}

			if param.(*ast.Parameter).AmpersandTkn != nil {
				name = "&" + name
			}

			removedParamNames = append(removedParamNames, name)
		}
	}

	return removedParamNames
}

func (e *ElementAvailableAttribute) shouldRemove(
	attrGroups []ast.Vertex,
) (bool, []ast.Vertex) {
	for attrI, attrGroup := range attrGroups {
	Attributes:
		for _, attr := range attrGroup.(*ast.AttributeGroup).Attrs {
			if len(attr.(*ast.Attribute).Args) == 0 {
				continue
			}

			if !e.targetter.MatchName(attr.(*ast.Attribute).Name) {
				continue
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
						return true, attrGroups
					}
				}

				if bytes.Equal(n, []byte("to")) || i == 1 {
					if e.version.IsHigherThan(v) {
						return true, attrGroups
					}
				}
			}

			e.logRemoval()
			attrGroups = slices.Delete(attrGroups, attrI, attrI+1)
			if len(attrGroups) == 0 {
				return false, nil
			}
			return false, attrGroups
		}
	}

	return false, attrGroups
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

	for _, modifier := range function.Modifiers {
		modifier.(*ast.Identifier).IdentifierTkn.FreeFloating = e.removeParamsDoc(
			modifier.(*ast.Identifier).IdentifierTkn.FreeFloating, params,
		)
	}

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

		doc, err := phpdoxer.ParseFullDoc(string(t.Value))
		if err != nil {
			log.Println(err)
			return freefloatings
		}

		newNodes := make([]phpdoxer.Node, 0, len(doc.Nodes))
		for _, node := range doc.Nodes {
			if paramNode, ok := node.(*phpdoxer.NodeParam); ok {
				if slices.Contains(params, paramNode.Name) {
					e.logRemoval()
					continue
				}
			}

			newNodes = append(newNodes, node)
		}

		doc.Nodes = newNodes
		t.Value = []byte(doc.String())
	}

	return freefloatings
}

func (e *ElementAvailableAttribute) logRemoval() {
	if e.logger != nil {
		e.logger.Printf("x")
	}
}
