package traversers

import (
	"errors"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
)

var ErrOnlyFQNSupported = errors.New(
	"Class definition traverser only supports fully qualified names",
)

func NewClassLike(name *ir.Name) (*ClassLike, error) {
	if !name.IsFullyQualified() {
		return nil, ErrOnlyFQNSupported
	}

	namespace := strings.Trim(strings.TrimSuffix(name.Value, name.LastPart()), partSeperator)

	return &ClassLike{
		fqn:                  name.Value,
		namespace:            namespace,
		class:                name.LastPart(),
		fileUses:             make([]*ir.UseStmt, 0),
		fileMatchesNamespace: true,
	}, nil
}

// ClassLike implements ir.Visitor.
type ClassLike struct {
	// Search query.
	fqn       string
	namespace string
	class     string

	// Context about the current file.
	fileUses             []*ir.UseStmt
	fileNamespace        string
	fileMatchesNamespace bool

	// Resulting found *ir.ClassStmt, *ir.InterfaceStmt or *ir.TraitStmt.
	Result ir.Node
}

func (c *ClassLike) EnterNode(node ir.Node) bool {
	// Already got our result.
	if c.Result != nil {
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

	var foundClassLikeNode ir.Node
	var foundClassLikeFQN string

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
		foundClassLikeFQN = c.nameToFQN(typedNode.ClassName.Value)
		foundClassLikeNode = typedNode
	case *ir.InterfaceStmt:
		foundClassLikeFQN = c.nameToFQN(typedNode.InterfaceName.Value)
		foundClassLikeNode = typedNode
	case *ir.TraitStmt:
		foundClassLikeFQN = c.nameToFQN(typedNode.TraitName.Value)
		foundClassLikeNode = typedNode
	}

	if foundClassLikeFQN != "" {
		if c.fqn == foundClassLikeFQN {
			c.Result = foundClassLikeNode
		}

		return false
	}

	return true
}

func (c *ClassLike) LeaveNode(ir.Node) {}

func (c *ClassLike) nameToFQN(name string) string {
	foundClassLike := c.fileNamespace + partSeperator + name
	if !strings.HasPrefix(foundClassLike, partSeperator) {
		foundClassLike = partSeperator + foundClassLike
	}

	return foundClassLike
}
