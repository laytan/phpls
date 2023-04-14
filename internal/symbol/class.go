package symbol

import (
	"fmt"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
	"github.com/laytan/phpls/internal/fqner"
	"github.com/laytan/phpls/internal/wrkspc"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/traversers"
)

type rooter interface {
	Root() *ast.Root
	Path() string
}

type fullyQualifier interface {
	GetFQN() *fqn.FQN
}

type ClassLike struct {
	*inheritor
	*modified
	*doxed

	rooter
	fullyQualifier

	node ast.Vertex
}

func NewClassLike(root rooter, node ast.Vertex) *ClassLike {
	qualifier := fqner.New(root, node)

	return &ClassLike{
		fullyQualifier: qualifier,
		inheritor: &inheritor{
			fullyQualifier: qualifier,
			rooter:         root,
		},
		modified: newModifiedFromNode(node),
		doxed:    NewDoxed(node),
		rooter:   root,
		node:     node,
	}
}

func NewClassLikeFromName(nameRoot *ast.Root, n ast.Vertex) (*ClassLike, error) {
	iNode, ok := fqner.FindFullyQualifiedName(nameRoot, n)
	if !ok {
		return nil, fmt.Errorf("[symbol.NewClassLikeFromName]: can't find %v in index", n)
	}

	root := wrkspc.Current.FIROf(iNode.Path)
	ct := traversers.NewClassLike(nodeident.Get(n))
	tv := traverser.NewTraverser(ct)
	root.Accept(tv)
	if ct.ClassLike == nil {
		return nil, fmt.Errorf(
			"[symbol.NewClassLikeFromName]: can't find class node for %s in %s",
			iNode.FQN,
			iNode.Path,
		)
	}

	return NewClassLike(wrkspc.NewRooter(iNode.Path, root), ct.ClassLike), nil
}

func NewClassLikeFromMethod(root *ast.Root, method *ast.StmtClassMethod) (*ClassLike, error) {
	napt := traversers.NewNodeAtPos(method.Position.StartPos)
	tv := traverser.NewTraverser(napt)
	root.Accept(tv)

	for i := len(napt.Nodes) - 1; i >= 0; i-- {
		switch node := napt.Nodes[i].(type) {
		case *ast.StmtClass:
			return NewClassLikeFromName(root, nodeToName(node.Name))
		case *ast.StmtTrait:
			return NewClassLikeFromName(root, nodeToName(node.Name))
		case *ast.StmtInterface:
			return NewClassLikeFromName(root, nodeToName(node.Name))
		}
	}

	return nil, fmt.Errorf(
		"[symbol.NewClassLikeFromMethod]: can't find class-like surrounding the given method %s",
		nodeident.Get(method.Name),
	)
}

func NewClassLikeFromProperty(root *ast.Root, property *ast.StmtPropertyList) (*ClassLike, error) {
	napt := traversers.NewNodeAtPos(property.Position.StartPos)
	tv := traverser.NewTraverser(napt)
	root.Accept(tv)

	for i := len(napt.Nodes) - 1; i >= 0; i-- {
		switch node := napt.Nodes[i].(type) {
		case *ast.StmtClass:
			return NewClassLikeFromName(root, nodeToName(node.Name))
		case *ast.StmtTrait:
			return NewClassLikeFromName(root, nodeToName(node.Name))
		}
	}

	return nil, fmt.Errorf("finding surrounding class of property: not found")
}

func NewClassLikeFromFQN(r rooter, qualified *fqn.FQN) (*ClassLike, error) {
	trav := traversers.NewClassLike(qualified.Name())
	tv := traverser.NewTraverser(trav)
	r.Root().Accept(tv)

	if trav.ClassLike == nil {
		return nil, fmt.Errorf(
			"symbol.NewClassLikeFromFQN: unable to find %v in given root",
			qualified,
		)
	}

	return NewClassLike(r, trav.ClassLike), nil
}

func (c *ClassLike) Kind() ast.Type {
	return c.node.GetType()
}

func (c *ClassLike) Name() string {
	return nodeident.Get(c.node)
}

type inheritor struct {
	fullyQualifier
	rooter

	traverser *inheritsTraverser
}

func (i *inheritor) ensureTraversed() {
	if i.traverser != nil {
		return
	}

	i.traverser = newInheritsTraverser(i.GetFQN())
	tv := traverser.NewTraverser(i.traverser)
	i.Root().Accept(tv)
}

func (i *inheritor) Uses() []ast.Vertex {
	i.ensureTraversed()
	return i.traverser.uses
}

func (i *inheritor) Extends() ast.Vertex {
	i.ensureTraversed()
	return i.traverser.extends
}

func (i *inheritor) Implements() []ast.Vertex {
	i.ensureTraversed()
	return i.traverser.implements
}
