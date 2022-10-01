package fqn

import (
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/symbol"
)

const partSeperator = `\`

type FQN struct {
	// Examples: \DateTime, \Test\DateTime.
	value string
}

func New(value string) *FQN {
	if value[0:1] != partSeperator {
		panic("Trying to create FQN without a fully qualified input.")
	}

	r := &FQN{value: value}
	return r
}

func (f *FQN) String() string {
	return f.value
}

func (f *FQN) Namespace() string {
	parts := strings.Split(f.value, partSeperator)
	return strings.Join(parts[1:len(parts)-1], partSeperator)
}

func (f *FQN) Name() string {
	parts := strings.Split(f.value, partSeperator)
	return parts[len(parts)-1]
}

func NewTraverser() *Traverser {
	return &Traverser{}
}

// A FQNTraverser that handles keywords like self or static.
func NewTraverserHandlingKeywords(block ir.Node) *Traverser {
	return &Traverser{block: block}
}

// Traverser implements ir.Visitor.
type Traverser struct {
	fileUses      []*ir.UseStmt
	fileNamespace string

	block      ir.Node
	blockClass *ir.ClassStmt
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

	// If any use statement ends with the class name, use that.
	for _, usage := range f.fileUses {
		if usage.Alias != nil {
			if usage.Alias.Value == name.LastPart() {
				return New(partSeperator + usage.Use.Value)
			}
		}

		if usage.Use.LastPart() == name.LastPart() {
			return New(partSeperator + usage.Use.Value)
		}
	}

	// Else use namespace+class name.
	if f.fileNamespace != "" {
		return New(partSeperator + f.fileNamespace + partSeperator + name.LastPart())
	}

	// Else use class name.
	return New(partSeperator + name.LastPart())
}

func (f *Traverser) EnterNode(node ir.Node) bool {
	switch typedNode := node.(type) {
	case *ir.ClassStmt:
		if f.block == nil {
			return false
		}

		bPos := ir.GetPosition(f.block)
		nPos := typedNode.Position

		// If the block is inside the class.
		if bPos.StartLine >= nPos.StartLine && bPos.EndLine <= nPos.EndLine && bPos.StartPos >= nPos.StartPos && bPos.EndPos <= nPos.EndPos {
			f.blockClass = typedNode
		}

		return false

	case *ir.UseStmt:
		f.fileUses = append(f.fileUses, typedNode)

		return false

	case *ir.NamespaceStmt:
		if name := symbol.GetIdentifier(typedNode); name != "" {
			f.fileNamespace = name
		}

		return false

	default:
		return !symbol.IsScope(typedNode)
	}
}

func (f *Traverser) LeaveNode(ir.Node) {}
