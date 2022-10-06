package phpdoxer

import (
	"fmt"
	"strings"

	"github.com/laytan/elephp/pkg/phpversion"
)

type NodeKind uint

const (
	KindUnknown NodeKind = iota
	KindReturn
	KindVar
	KindParam
	KindInheritDoc
	KindSince
    KindRemoved
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

type NodeParam struct {
	Type Type
	Name string
}

func (n *NodeParam) String() string {
	return fmt.Sprintf("@param %s %s", n.Type.String(), n.Name)
}

func (n *NodeParam) Kind() NodeKind {
	return KindParam
}

type NodeInheritDoc struct{}

func (n *NodeInheritDoc) String() string {
	return "{@inheritdoc}"
}

func (n *NodeInheritDoc) Kind() NodeKind {
	return KindInheritDoc
}

type NodeSince struct {
	Version     *phpversion.PHPVersion
	Description string
}

func (n *NodeSince) String() string {
	return strings.TrimSpace("@since " + n.Version.String() + " " + n.Description)
}

func (n *NodeSince) Kind() NodeKind {
	return KindSince
}

type NodeRemoved struct {
	Version     *phpversion.PHPVersion
	Description string
}

func (n *NodeRemoved) String() string {
	return strings.TrimSpace("@removed " + n.Version.String() + " " + n.Description)
}

func (n *NodeRemoved) Kind() NodeKind {
	return KindRemoved
}
