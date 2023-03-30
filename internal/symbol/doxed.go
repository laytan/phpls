package symbol

import (
	"fmt"
	"log"
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/token"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/phpdoxer"
)

type DocFilter func(phpdoxer.Node) bool

func FilterDocKind(kind phpdoxer.NodeKind) DocFilter {
	return func(n phpdoxer.Node) bool {
		return n.Kind() == kind
	}
}

type doxed struct {
	node ast.Vertex

	docNodeCache []phpdoxer.Node
}

func NewDoxed(
	node ast.Vertex,
) *doxed { //nolint:revive // This struct is embedded inside other symbol structs, if we leave it public, it will be accessible by users.
	return &doxed{node: node}
}

func (d *doxed) Docs() []phpdoxer.Node {
	if d.docNodeCache != nil {
		return d.docNodeCache
	}

	cmnts := NodeComments(d.node)
	nodes := make([]phpdoxer.Node, 0, len(cmnts))
	for _, cmnt := range cmnts {
		cmntNodes, err := phpdoxer.ParseDoc(cmnt)
		if err != nil {
			log.Println(fmt.Errorf("[symbol.Doxed.Docs]: err parsing doc \"%s\": %w", cmnt, err))
			continue
		}

		nodes = append(nodes, cmntNodes...)
	}

	d.docNodeCache = nodes
	return nodes
}

func (d *doxed) FindDoc(filters ...DocFilter) phpdoxer.Node {
	docs := d.Docs()

	for _, doc := range docs {
		isValid := true
		for _, filter := range filters {
			if !filter(doc) {
				isValid = false
				break
			}
		}

		if isValid {
			return doc
		}
	}

	return nil
}

func (d *doxed) FindAllDocs(filters ...DocFilter) (results []phpdoxer.Node) {
	docs := d.Docs()

DocsRange:
	for _, doc := range docs {
		for _, filter := range filters {
			if !filter(doc) {
				break DocsRange
			}
		}

		results = append(results, doc)
	}

	return results
}

func (d *doxed) FindThrows() []*phpdoxer.NodeThrows {
	results := d.FindAllDocs(FilterDocKind(phpdoxer.KindThrows))
	throws := make([]*phpdoxer.NodeThrows, 0, len(results))
	throws = append(throws, functional.Map(results, func(n phpdoxer.Node) *phpdoxer.NodeThrows {
		typed, ok := n.(*phpdoxer.NodeThrows)
		if !ok {
			log.Panic(
				"[doxed.FindThrows]: phpdoxer node with kind phpdoxer.KindThrows is not of type *phpdoxer.NodeThrows",
			)
		}

		return typed
	})...)

	return throws
}

func (d *doxed) RawDocs() string {
	return strings.Join(NodeComments(d.node), "\n")
}

// TypeHintToDocType Turns a type hint node, from for example ir.FunctionStmt.ReturnType into the
// equivalent phpdoxer.Type.
func TypeHintToDocType(node ast.Vertex) (phpdoxer.Type, error) {
	var isNullable bool
	if nullable, ok := node.(*ast.Nullable); ok {
		node = nullable.Expr
		isNullable = true
	}

	name, ok := node.(*ast.Name)
	if !ok {
		return nil, fmt.Errorf(
			"%T is unsupported for a return type hint, expecting *ir.Name or *ir.Nullable",
			node,
		)
	}

	toParse := nodeident.Get(name)
	if isNullable {
		toParse = "null|" + toParse
	}

	t, err := phpdoxer.ParseType(toParse)
	if err != nil {
		return nil, fmt.Errorf(`parsing type hint into doc type "%s": %w`, toParse, err)
	}

	return t, nil
}

func NodeComments(node ast.Vertex) []string {
	var ff []*token.Token
	switch tn := node.(type) {
	case *ast.StmtFunction:
		if len(tn.AttrGroups) > 0 {
			ff = tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating
			break
		}

		ff = tn.FunctionTkn.FreeFloating
	case *ast.ExprArrowFunction:
		if len(tn.AttrGroups) > 0 {
			ff = tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating
			break
		}

		if tn.StaticTkn != nil {
			ff = tn.StaticTkn.FreeFloating
			break
		}

		ff = tn.FnTkn.FreeFloating
	case *ast.StmtTrait:
		if len(tn.AttrGroups) > 0 {
			ff = tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating
			break
		}

		ff = tn.TraitTkn.FreeFloating
	case *ast.StmtClass:
		if len(tn.AttrGroups) > 0 {
			ff = tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating
			break
		}

		if len(tn.Modifiers) > 0 {
			ff = tn.Modifiers[0].(*ast.Identifier).IdentifierTkn.FreeFloating
			break
		}

		ff = tn.ClassTkn.FreeFloating
	case *ast.StmtInterface:
		if len(tn.AttrGroups) > 0 {
			ff = tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating
			break
		}

		ff = tn.InterfaceTkn.FreeFloating
	case *ast.StmtPropertyList:
		if len(tn.AttrGroups) > 0 {
			ff = tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating
			break
		}

		if len(tn.Modifiers) > 0 {
			ff = tn.Modifiers[0].(*ast.Identifier).IdentifierTkn.FreeFloating
			break
		}

		log.Println("Warning: *ast.StmtPropertyList without any attributes or modifiers, is that possible?")
	case *ast.StmtClassMethod:
		if len(tn.AttrGroups) > 0 {
			ff = tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating
			break
		}

		if len(tn.Modifiers) > 0 {
			ff = tn.Modifiers[0].(*ast.Identifier).IdentifierTkn.FreeFloating
			break
		}

		ff = tn.FunctionTkn.FreeFloating
	case *ast.ExprClosure:
		if len(tn.AttrGroups) > 0 {
			ff = tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating
			break
		}

		if tn.StaticTkn != nil {
			ff = tn.StaticTkn.FreeFloating
			break
		}

		ff = tn.FunctionTkn.FreeFloating
	case *ast.StmtClassConstList:
		if len(tn.AttrGroups) > 0 {
			ff = tn.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating
			break
		}

		if len(tn.Modifiers) > 0 {
			ff = tn.Modifiers[0].(*ast.Identifier).IdentifierTkn.FreeFloating
			break
		}

		ff = tn.ConstTkn.FreeFloating
	}

	docs := []string{}
	for _, f := range ff {
		if f.ID != token.T_COMMENT && f.ID != token.T_DOC_COMMENT {
			continue
		}

		docs = append(docs, string(f.Value))
	}

	return docs
}
