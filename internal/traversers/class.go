package traversers

import (
	"errors"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
)

var ErrOnlyFQNSupported = errors.New(
	"Class definition traverser only supports fully qualified names",
)

func NewClass(name *ir.Name) (*Class, error) {
	if !name.IsFullyQualified() {
		return nil, ErrOnlyFQNSupported
	}

	namespace := strings.Trim(strings.TrimSuffix(name.Value, name.LastPart()), partSeperator)

	return &Class{
		fqn:                  name.Value,
		namespace:            namespace,
		class:                name.LastPart(),
		fileUses:             make([]*ir.UseStmt, 0),
		fileMatchesNamespace: true,
	}, nil
}

// Class implements ir.Visitor.
type Class struct {
	// Search query.
	fqn       string
	namespace string
	class     string

	// Context about the current file.
	fileUses             []*ir.UseStmt
	fileNamespace        string
	fileMatchesNamespace bool

	// Resulting found class.
	ResultClass *ir.ClassStmt
}

func (c *Class) EnterNode(node ir.Node) bool {
	// Already got our result.
	if c.ResultClass != nil {
		return false
	}

	// If the current file does not match the namespace,
	// and we are not the root node, there is no reason to check this node.
	if c.fileNamespace != "" && c.fileNamespace != c.namespace {
		switch node.(type) {
		case *ir.Root:
			break
		default:
			return false
		}
	}

	switch typedNode := node.(type) {

	// Reset the file state because the root node signals a new file.
	case *ir.Root:
		c.fileUses = make([]*ir.UseStmt, 0)
		c.fileNamespace = ""
		c.fileMatchesNamespace = true

	case *ir.UseStmt:
		c.fileUses = append(c.fileUses, typedNode)
		return false

	case *ir.NamespaceStmt:
		if typedNode.NamespaceName == nil {
			return false
		}

		c.fileNamespace = typedNode.NamespaceName.Value
		return false

	case *ir.ClassStmt:
		foundClass := c.fileNamespace + partSeperator + typedNode.ClassName.Value
		if !strings.HasPrefix(foundClass, partSeperator) {
			foundClass = partSeperator + foundClass
		}

		if c.fqn == foundClass {
			c.ResultClass = typedNode
		}

		return false
	}

	return true
}

func (c *Class) LeaveNode(ir.Node) {}
