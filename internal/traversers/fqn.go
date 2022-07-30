package traversers

import (
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/symbol"
)

type FQN struct {
	// Examples: \DateTime, \Test\DateTime.
	value string
}

func NewFQN(value string) *FQN {
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

func NewFQNTraverser(name *ir.Name) *FQNTraverser {
	return &FQNTraverser{name: name}
}

// FQNTraverser implements ir.Visitor.
type FQNTraverser struct {
	name *ir.Name

	// Context about the current file.
	fileUses      []*ir.UseStmt
	fileNamespace string
}

func (f *FQNTraverser) Result() *FQN {
	// If any use statement ends with the class name, use that.
	for _, usage := range f.fileUses {
		if usage.Alias != nil {
			if usage.Alias.Value == f.name.LastPart() {
				return NewFQN(partSeperator + usage.Use.Value)
			}
		}

		if usage.Use.LastPart() == f.name.LastPart() {
			return NewFQN(partSeperator + usage.Use.Value)
		}
	}

	// Else use namespace+class name.
	if f.fileNamespace != "" {
		return NewFQN(partSeperator + f.fileNamespace + partSeperator + f.name.LastPart())
	}

	// Else use class name.
	return NewFQN(partSeperator + f.name.LastPart())
}

func (f *FQNTraverser) EnterNode(node ir.Node) bool {
	switch typedNode := node.(type) {
	case *ir.Root, *ir.UseListStmt:
		return true

	case *ir.UseStmt:
		f.fileUses = append(f.fileUses, typedNode)

		return false

	case *ir.NamespaceStmt:
		if name := symbol.GetIdentifier(typedNode); name != "" {
			f.fileNamespace = name
		}

		return false

	default:
		return false
	}
}

func (f *FQNTraverser) LeaveNode(ir.Node) {}
