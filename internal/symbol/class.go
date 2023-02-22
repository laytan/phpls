package symbol

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/fqner"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/traversers"
)

type rooter interface {
	Root() *ir.Root
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

	node ir.Node
}

func NewClassLike(root rooter, node ir.Node) *ClassLike {
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

func NewClassLikeFromName(nameRoot *ir.Root, n *ir.Name) (*ClassLike, error) {
	iNode, ok := fqner.FindFullyQualifiedName(nameRoot, n)
	if !ok {
		return nil, fmt.Errorf("[symbol.NewClassLikeFromName]: can't find %v in index", n)
	}

	root := wrkspc.FromContainer().FIROf(iNode.Path)
	ct := traversers.NewClassLike(n.Value)
	root.Walk(ct)
	if ct.ClassLike == nil {
		return nil, fmt.Errorf(
			"[symbol.NewClassLikeFromName]: can't find class node for %s in %s",
			iNode.FQN,
			iNode.Path,
		)
	}

	return NewClassLike(wrkspc.NewRooter(iNode.Path, root), ct.ClassLike), nil
}

func NewClassLikeFromMethod(root *ir.Root, method *ir.ClassMethodStmt) (*ClassLike, error) {
	napt := traversers.NewNodeAtPos(uint(method.Position.StartPos))
	root.Walk(napt)

	for i := len(napt.Nodes) - 1; i >= 0; i-- {
		switch node := napt.Nodes[i].(type) {
		case *ir.ClassStmt:
			return NewClassLikeFromName(root, nodeToName(node.ClassName))
		case *ir.TraitStmt:
			return NewClassLikeFromName(root, nodeToName(node.TraitName))
		case *ir.InterfaceStmt:
			return NewClassLikeFromName(root, nodeToName(node.InterfaceName))
		}
	}

	return nil, fmt.Errorf(
		"[symbol.NewClassLikeFromMethod]: can't find class-like surrounding the given method %s",
		method.MethodName.Value,
	)
}

func NewClassLikeFromFQN(r rooter, qualified *fqn.FQN) (*ClassLike, error) {
	trav := traversers.NewClassLike(qualified.Name())
	r.Root().Walk(trav)

	if trav.ClassLike == nil {
		return nil, fmt.Errorf(
			"symbol.NewClassLikeFromFQN: unable to find %v in given root",
			qualified,
		)
	}

	return NewClassLike(r, trav.ClassLike), nil
}

func (c *ClassLike) Kind() ir.NodeKind {
	return ir.GetNodeKind(c.node)
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
	i.Root().Walk(i.traverser)
}

func (i *inheritor) Uses() []*ir.Name {
	i.ensureTraversed()
	return i.traverser.uses
}

func (i *inheritor) Extends() *ir.Name {
	i.ensureTraversed()
	return i.traverser.extends
}

func (i *inheritor) Implements() []*ir.Name {
	i.ensureTraversed()
	return i.traverser.implements
}
