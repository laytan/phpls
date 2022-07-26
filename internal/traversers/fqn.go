package traversers

import (
	"github.com/VKCOM/noverify/src/ir"
)

func NewFQN(name *ir.Name) (*FQN, error) {
	return &FQN{
		name: name,
	}, nil
}

// FQN implements ir.Visitor.
type FQN struct {
	name *ir.Name

	// Context about the current file.
	fileUses      []*ir.UseStmt
	fileNamespace string
}

func (f *FQN) Result() string {
	// If any use statement ends with the class name, use that.
	for _, usage := range f.fileUses {
		if usage.Alias != nil {
			if usage.Alias.Value == f.name.LastPart() {
				return partSeperator + usage.Use.Value
			}
		}

		if usage.Use.LastPart() == f.name.LastPart() {
			return partSeperator + usage.Use.Value
		}
	}

	// Else use namespace+class name.
	if f.fileNamespace != "" {
		return partSeperator + f.fileNamespace + partSeperator + f.name.LastPart()
	}

	// Else use class name.
	return partSeperator + f.name.LastPart()
}

func (f *FQN) EnterNode(node ir.Node) bool {
	switch typedNode := node.(type) {

	case *ir.UseStmt:
		f.fileUses = append(f.fileUses, typedNode)
		return false

	case *ir.NamespaceStmt:
		if typedNode.NamespaceName == nil {
			return false
		}

		f.fileNamespace = typedNode.NamespaceName.Value
		return false
	}

	return true
}

func (f *FQN) LeaveNode(ir.Node) {}
