package symbol

import (
	"log"

	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/set"
	"github.com/laytan/php-parser/pkg/ast"
)

type Modified interface {
	Privacy() phprivacy.Privacy
	CanBeAccessedFrom(phprivacy.Privacy) bool
	IsFinal() bool
	IsStatic() bool
}

func FilterPrivacy[T Modified](privacy phprivacy.Privacy) FilterFunc[T] {
	return func(v T) bool {
		return v.Privacy() == privacy
	}
}

func FilterNotPrivacy[T Modified](privacy phprivacy.Privacy) FilterFunc[T] {
	return func(v T) bool {
		return v.Privacy() != privacy
	}
}

func FilterCanBeAccessedFrom[T Modified](privacy phprivacy.Privacy) FilterFunc[T] {
	return func(v T) bool {
		return v.CanBeAccessedFrom(privacy)
	}
}

func FilterStatic[T Modified]() FilterFunc[T] {
	return func(v T) bool {
		return v.IsStatic()
	}
}

func FilterNotStatic[T Modified]() FilterFunc[T] {
	return func(v T) bool {
		return !v.IsStatic()
	}
}

type modified struct {
	modifiers *set.Set[string]
}

func newModified(modifiers *set.Set[string]) *modified {
	return &modified{
		modifiers: modifiers,
	}
}

func newModifiedFromNode(node ast.Vertex) *modified {
	var mods []ast.Vertex
	switch typedNode := node.(type) {
	case *ast.StmtClassMethod:
		mods = typedNode.Modifiers
	case *ast.StmtClass:
		mods = typedNode.Modifiers
	case *ast.StmtPropertyList:
		mods = typedNode.Modifiers
	case *ast.StmtClassConstList:
		mods = typedNode.Modifiers
	case *ast.StmtTrait, *ast.StmtInterface:
		// An interface or trait is a valid class-like but never has modifiers,
		// even though, lets not complain using the default case.
		mods = []ast.Vertex{}
	default:
		log.Printf("[symbol.NewModifiedFromNode]: %T is not a node that can be modified", node)
		mods = []ast.Vertex{}
	}

	return newModified(set.NewFromSlice(functional.Map(mods, nodeident.Get)))
}

func (m *modified) Privacy() phprivacy.Privacy {
	if m.modifiers.Has("private") {
		return phprivacy.PrivacyPrivate
	}

	if m.modifiers.Has("protected") {
		return phprivacy.PrivacyProtected
	}

	// A symbol is public unless otherwise specified.
	return phprivacy.PrivacyPublic
}

func (m *modified) CanBeAccessedFrom(p phprivacy.Privacy) bool {
	return p.CanAccess(m.Privacy())
}

func (m *modified) IsStatic() bool {
	return m.modifiers.Has("static")
}

func (m *modified) IsFinal() bool {
	return m.modifiers.Has("final")
}
