package fqn

import (
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/position"
	"github.com/laytan/elephp/pkg/symbol"
)

// Traverser implements ir.Visitor.
type Traverser struct {
	namespaces []*namespace

	block      ir.Node
	blockClass *ir.ClassStmt
}

func NewTraverser() *Traverser {
	return &Traverser{namespaces: []*namespace{globalNamespace()}}
}

// A FQNTraverser that handles keywords like self or static.
func NewTraverserHandlingKeywords(block ir.Node) *Traverser {
	return &Traverser{block: block, namespaces: []*namespace{globalNamespace()}}
}

func (f *Traverser) ResultFor(name *ir.Name) *FQN {
	// Handle self and static by returning the class the block is in.
	if f.block != nil && f.blockClass != nil {
		if name.Value == "self" || name.Value == "static" {
			name.Value = f.blockClass.ClassName.Value
		}
	}

	if name.IsFullyQualified() {
		return New(name.Value)
	}

	if name.Position == nil {
		log.Println(fmt.Errorf(
			"[fqn.Traverser.ResultFor(%v)]: given a name without a position is discouraged because it could return wrong results if the file has multiple namespaces defined",
			name,
		))
	}

	// Default to first namespace.
	ns := f.namespaces[0]

	// If no position is defined, use the last namespace as the namespace.
	// NOTE: This can be wrong if there are multiple namespaces defined in the file though.
	if name.Position == nil {
		ns = f.namespaces[len(f.namespaces)-1]
	}

	if name.Position != nil {
		// If any other namespace contains the given name, use that one.
		for _, possibleNs := range f.namespaces {
			if possibleNs.Contains(name.Position) {
				ns = possibleNs
				break
			}
		}
	}

	// If any use statement ends with the class name, use that.
	for _, usage := range ns.uses {
		if usage.Alias != nil {
			if usage.Alias.Value == name.LastPart() {
				return New(PartSeperator + usage.Use.Value)
			}
		}

		if usage.Use.LastPart() == name.LastPart() {
			return New(PartSeperator + usage.Use.Value)
		}
	}

	// Else use namespace+class name.
	if ns.ns != "" {
		return New(PartSeperator + ns.ns + PartSeperator + name.LastPart())
	}

	// Else use class name.
	return New(PartSeperator + name.LastPart())
}

func (f *Traverser) EnterNode(node ir.Node) bool {
	switch typedNode := node.(type) {
	case *ir.ClassStmt:
		if f.block == nil {
			return false
		}

		bPos := ir.GetPosition(f.block)
		nPos := typedNode.Position

		if posContainsPos(nPos, bPos) {
			f.blockClass = typedNode
		}

		return false

	case *ir.UseStmt:
		currNs := f.namespaces[len(f.namespaces)-1]
		currNs.uses = append(currNs.uses, typedNode)

		return false

	case *ir.NamespaceStmt:
		ns := &namespace{
			ns:   symbol.GetIdentifier(typedNode),
			uses: []*ir.UseStmt{},
			pos:  &position.Position{},
		}

		if typedNode.SemiColonTkn != nil {
			ns.pos.StartLine = typedNode.SemiColonTkn.Position.StartLine
			ns.pos.StartPos = typedNode.SemiColonTkn.Position.StartPos + 1
		}

		if typedNode.OpenCurlyBracketTkn != nil {
			ns.pos.StartLine = typedNode.OpenCurlyBracketTkn.Position.StartLine
			ns.pos.StartPos = typedNode.OpenCurlyBracketTkn.Position.StartPos + 1
		}

		if typedNode.CloseCurlyBracketTkn != nil {
			ns.pos.EndLine = typedNode.CloseCurlyBracketTkn.Position.EndLine
			ns.pos.EndPos = typedNode.CloseCurlyBracketTkn.Position.EndPos - 1
		}

		// If the last namespace was a non-curlybraced namespace (no end), its end
		// is the start of this namespace.
		prevNs := f.namespaces[len(f.namespaces)-1]
		if prevNs.pos.EndLine == 0 && prevNs.pos.EndPos == 0 {
			prevNs.pos.EndLine = typedNode.NsTkn.Position.StartLine
			prevNs.pos.EndPos = typedNode.NsTkn.Position.StartPos - 1
		}

		f.namespaces = append(f.namespaces, ns)

		return true

	default:
		return !symbol.IsScope(ir.GetNodeKind(typedNode))
	}
}

func (f *Traverser) LeaveNode(ir.Node) {}

type namespace struct {
	ns   string
	uses []*ir.UseStmt
	pos  *position.Position
}

func (n *namespace) Contains(pos *position.Position) bool {
	return posContainsPos(n.pos, pos)
}

func posContainsPos(pos *position.Position, pos2 *position.Position) bool {
	// If it has no end position, act as if end is infinity.
	if pos.EndLine == 0 && pos.EndPos == 0 {
		return pos2.StartLine >= pos.StartLine && pos2.StartPos >= pos.StartPos
	}

	return pos2.StartLine >= pos.StartLine && pos2.EndLine <= pos.EndLine &&
		pos2.StartPos >= pos.StartPos &&
		pos2.EndPos <= pos.EndPos
}

func globalNamespace() *namespace {
	return &namespace{
		ns:   "",
		uses: []*ir.UseStmt{},
		pos:  &position.Position{},
	}
}
