package visitor

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/laytan/elephp/pkg/phpversion"
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
				typedStmt.Params = e.filterParams(typedStmt.Params)
			}

		case *ast.StmtClassMethod:
			ok = !e.shouldRemove(typedStmt.AttrGroups)
			if ok {
				typedStmt.Params = e.filterParams(typedStmt.Params)
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

func (e *ElementAvailableAttribute) filterParams(params []ast.Vertex) []ast.Vertex {
	newParams := make([]ast.Vertex, 0, len(params))

	for _, param := range params {
		if typedParam, ok := param.(*ast.Parameter); ok {
			if !e.shouldRemove(typedParam.AttrGroups) {
				newParams = append(newParams, param)
				continue
			}

			e.logRemoval(param)
		}
	}

	return newParams
}

func (e *ElementAvailableAttribute) shouldRemove(attrGroups []ast.Vertex) bool {
	for _, attrGroup := range attrGroups {
		if typedAttrGroup, ok := attrGroup.(*ast.AttributeGroup); ok {
		Attributes:
			for _, attr := range typedAttrGroup.Attrs {
				if typedAttr, ok := attr.(*ast.Attribute); ok {
					if len(typedAttr.Args) == 0 {
						continue
					}

					if attrName, ok := typedAttr.Name.(*ast.Name); ok {
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
					}

					for i, arg := range typedAttr.Args {
						if i > 1 {
							break
						}

						var n []byte
						var v *phpversion.PHPVersion
						if typedArg, ok := arg.(*ast.Argument); ok {
							if argName, ok := typedArg.Name.(*ast.Identifier); ok {
								n = argName.Value
							}

							if exprStr, ok := typedArg.Expr.(*ast.ScalarString); ok {
								versionStr := strings.Trim(string(exprStr.Value), `'"`)
								if version, ok := phpversion.FromString(versionStr); ok {
									v = version
								}
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
		}
	}

	return false
}

func (e *ElementAvailableAttribute) logRemoval(n ast.Vertex) {
	if e.logging {
		_, _ = fmt.Printf("x") //nolint:forbidigo // For visualization.
	}
}
