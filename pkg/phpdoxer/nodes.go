package phpdoxer

import (
	"fmt"
	"strings"

	"github.com/laytan/phpls/pkg/phpversion"
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
	KindThrows
)

type Node interface {
	String() string
	Kind() NodeKind
	Range() (start int, end int)
	setRange(start int, end int)
}

type NodeRange struct {
	StartPos int
	EndPos   int
}

func (n *NodeRange) Range() (start int, end int) {
	return n.StartPos, n.EndPos
}

func (n *NodeRange) setRange(start int, end int) {
	n.StartPos = start
	n.EndPos = end
}

type NodeUnknown struct {
	NodeRange

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
	NodeRange

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
	NodeRange

	Type        Type
	Description string
}

func (n *NodeVar) String() string {
	if n.Description == "" {
		return fmt.Sprintf("@var %s", n.Type)
	}

	return fmt.Sprintf("@var %s %s", n.Type, n.Description)
}

func (n *NodeVar) Kind() NodeKind {
	return KindVar
}

type NodeParam struct {
	NodeRange

	Type        Type
	Name        string
	Description string
}

func (n *NodeParam) String() string {
	desc := n.Description
	if desc != "" {
		desc = " " + desc
	}

	if n.Type != nil {
		return fmt.Sprintf("@param %s %s%s", n.Type.String(), n.Name, desc)
	}

	return fmt.Sprintf("@param %s%s", n.Name, desc)
}

func (n *NodeParam) Kind() NodeKind {
	return KindParam
}

type NodeInheritDoc struct {
	NodeRange
}

func (n *NodeInheritDoc) String() string {
	return "{@inheritdoc}"
}

func (n *NodeInheritDoc) Kind() NodeKind {
	return KindInheritDoc
}

type NodeSince struct {
	NodeRange

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
	NodeRange

	Version     *phpversion.PHPVersion
	Description string
}

func (n *NodeRemoved) String() string {
	return strings.TrimSpace("@removed " + n.Version.String() + " " + n.Description)
}

func (n *NodeRemoved) Kind() NodeKind {
	return KindRemoved
}

type NodeThrows struct {
	NodeRange

	Type        Type
	Description string
}

func (n *NodeThrows) String() string {
	if n.Description == "" {
		return fmt.Sprintf("@throws %s", n.Type)
	}

	return fmt.Sprintf("@throws %s %s", n.Type, n.Description)
}

func (n *NodeThrows) Kind() NodeKind {
	return KindThrows
}
