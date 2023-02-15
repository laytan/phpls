package symbol

import (
	"fmt"
	"log"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/typer"
)

type DocFilter func(phpdoxer.Node) bool

func FilterDocKind(kind phpdoxer.NodeKind) DocFilter {
	return func(n phpdoxer.Node) bool {
		return n.Kind() == kind
	}
}

type doxed struct {
	node ir.Node

	docNodeCache []phpdoxer.Node
}

func NewDoxed(
	node ir.Node,
) *doxed { //nolint:revive // This struct is embedded inside other symbol structs, if we leave it public, it will be accessible by users.
	return &doxed{node: node}
}

func (d *doxed) Docs() []phpdoxer.Node {
	if d.docNodeCache != nil {
		return d.docNodeCache
	}

	cmnts := typer.NodeComments(d.node)
	nodes := make([]phpdoxer.Node, 0, len(cmnts))
	for _, cmnt := range cmnts {
		cmntNodes, err := phpdoxer.ParseDoc(cmnt)
		if err != nil {
			log.Println(fmt.Errorf("[symbol.Doxed.Docs]: err parsing doc \"%s\": %w", cmnt, err))
			continue
		}

		nodes = append(nodes, cmntNodes...)
	}

	d.docNodeCache = nodes
	return nodes
}

func (d *doxed) FindDoc(filters ...DocFilter) phpdoxer.Node {
	docs := d.Docs()

	for _, doc := range docs {
		isValid := true
		for _, filter := range filters {
			if !filter(doc) {
				isValid = false
				break
			}
		}

		if isValid {
			return doc
		}
	}

	return nil
}

func (d *doxed) FindAllDocs(filters ...DocFilter) (results []phpdoxer.Node) {
	docs := d.Docs()

DocsRange:
	for _, doc := range docs {
		for _, filter := range filters {
			if !filter(doc) {
				break DocsRange
			}
		}

		results = append(results, doc)
	}

	return results
}

func (d *doxed) FindThrows() []*phpdoxer.NodeThrows {
	results := d.FindAllDocs(FilterDocKind(phpdoxer.KindThrows))
	throws := make([]*phpdoxer.NodeThrows, 0, len(results))
	throws = append(throws, functional.Map(results, func(n phpdoxer.Node) *phpdoxer.NodeThrows {
		typed, ok := n.(*phpdoxer.NodeThrows)
		if !ok {
			log.Panic(
				"[doxed.FindThrows]: phpdoxer node with kind phpdoxer.KindThrows is not of type *phpdoxer.NodeThrows",
			)
		}

		return typed
	})...)

	return throws
}

func (d *doxed) RawDocs() string {
	return strings.Join(typer.NodeComments(d.node), "\n")
}
