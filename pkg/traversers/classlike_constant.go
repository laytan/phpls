package traversers

import (
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/symbol"
)

func NewClassLikeConstant(
	name string,
	classLikeName string,
	privacy phprivacy.Privacy,
) *ClassLikeConstant {
	return &ClassLikeConstant{
		name:          name,
		classLikeName: classLikeName,
		privacy:       privacy,
	}
}

// ClassLikeConstant implements ir.Visitor.
type ClassLikeConstant struct {
	ClassLikeConstant *ir.ClassConstListStmt
	name              string
	classLikeName     string
	privacy           phprivacy.Privacy
}

func (m *ClassLikeConstant) EnterNode(node ir.Node) bool {
	if m.ClassLikeConstant != nil {
		return false
	}

	switch typedNode := node.(type) {
	// Only parse a class-like node if the name matches (for multiple classes in a file).
	case *ir.ClassStmt, *ir.InterfaceStmt, *ir.TraitStmt:
		return symbol.GetIdentifier(node) == m.classLikeName

	case *ir.ClassConstListStmt:
		hasConst := false
		for _, constNode := range typedNode.Consts {
			typedConstNode, ok := constNode.(*ir.ConstantStmt)
			if !ok {
				log.Printf("Unexpected node type %T for constant, expecting *ir.ConstantStmt", constNode)
				continue
			}

			if typedConstNode.ConstantName.Value == m.name {
				hasConst = true
			}
		}

		if !hasConst {
			return false
		}

		hasPrivacy := false
		for _, mod := range typedNode.Modifiers {
			privacy, err := phprivacy.FromString(mod.Value)
			if err != nil {
				continue
			}

			hasPrivacy = true

			if !m.privacy.CanAccess(privacy) {
				continue
			}

			m.ClassLikeConstant = typedNode
			return false
		}

		// No privacy, means public privacy, means accessible.
		if !hasPrivacy {
			m.ClassLikeConstant = typedNode
			return false
		}
	}

	return !symbol.IsScope(node)
}

func (m *ClassLikeConstant) LeaveNode(ir.Node) {}
