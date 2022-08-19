package phpdoxer

import "fmt"

type NodeKind uint

const (
	KindUnknown NodeKind = iota
	KindReturn
	KindVar
)

type Node interface {
	String() string
	Kind() NodeKind
}

type NodeUnknown struct {
	// The string after @.
	At    string
	Value string
}

func (n *NodeUnknown) String() string {
	return fmt.Sprintf("@%s %s", n.At, n.Value)
}

func (n *NodeUnknown) Kind() NodeKind {
	return KindUnknown
}

type NodeReturn struct {
	Type        Type
	Description string
}

func (n *NodeReturn) String() string {
	if n.Description == "" {
		return fmt.Sprintf("@return %s", n.Type)
	}

	return fmt.Sprintf("@return %s %s", n.Type, n.Description)
}

func (n *NodeReturn) Kind() NodeKind {
	return KindReturn
}

type NodeVar struct {
	Type Type
}

func (n *NodeVar) String() string {
	return fmt.Sprintf("@var %s", n.Type)
}

func (n *NodeVar) Kind() NodeKind {
	return KindVar
}
