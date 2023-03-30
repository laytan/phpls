package fqn

import (
	"strings"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/position"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/nodescopes"
)

type Traverser struct {
	visitor.Null
	namespaces []*namespace

	block      ast.Vertex
	blockClass *ast.StmtClass
}

func NewTraverser() *Traverser {
	return &Traverser{namespaces: []*namespace{globalNamespace()}}
}

// A FQNTraverser that handles keywords like self or static.
func NewTraverserHandlingKeywords(block ast.Vertex) *Traverser {
	return &Traverser{block: block, namespaces: []*namespace{globalNamespace()}}
}

func (f *Traverser) ResultFor2(position *position.Position, name string) *FQN {
	return f.ResultFor(&ast.Name{
		Position: position,
		Parts: functional.Map(
			strings.Split(name, "\\"),
			func(s string) ast.Vertex { return &ast.NamePart{Value: []byte(s)} },
		),
	})
}

func (f *Traverser) ResultFor(name ast.Vertex) *FQN {
	nv := nodeident.Get(name)
	// Handle self and static by returning the class the block is in.
	if f.block != nil && f.blockClass != nil {
		if nv == "self" || nv == "static" {
			nv = nodeident.Get(f.blockClass)
		}
	}

	if nv[0] == '\\' {
		return New(nv)
	}

	// Default to first namespace.
	ns := f.namespaces[0]

	// If any other namespace contains the given name, use that one.
	for _, possibleNs := range f.namespaces {
		if possibleNs.Contains(name.GetPosition()) {
			ns = possibleNs
			break
		}
	}

	parts := strings.Split(nv, "\\")
	cn := parts[len(parts)-1]

	// If any use statement ends with the class name, use that.
	for _, usage := range ns.uses {
		useIdent := nodeident.Get(usage.Use)
		if usage.Alias != nil {
			if nodeident.Get(usage.Alias) == cn {
				return New(PartSeperator + useIdent)
			}
		}

		useParts := strings.Split(useIdent, "\\")
		un := useParts[len(useParts)-1]
		if un == cn {
			return New(PartSeperator + useIdent)
		}
	}

	// Else use namespace+class name.
	if ns.ns != "" {
		return New(PartSeperator + ns.ns + PartSeperator + cn)
	}

	// Else use class name.
	return New(PartSeperator + cn)
}

func (f *Traverser) EnterNode(node ast.Vertex) bool {
	switch typedNode := node.(type) {
	case *ast.StmtClass:
		if f.block == nil {
			return false
		}

		bPos := f.block.GetPosition()
		nPos := typedNode.Position

		if posContainsPos(nPos, bPos) {
			f.blockClass = typedNode
		}

		return false

	case *ast.StmtUse:
		currNs := f.namespaces[len(f.namespaces)-1]
		currNs.uses = append(currNs.uses, typedNode)

		return false

	case *ast.StmtNamespace:
		ns := &namespace{
			ns:   nodeident.Get(typedNode),
			uses: []*ast.StmtUse{},
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
		return !nodescopes.IsScope(typedNode.GetType())
	}
}

type namespace struct {
	ns   string
	uses []*ast.StmtUse
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
		uses: []*ast.StmtUse{},
		pos:  &position.Position{},
	}
}
