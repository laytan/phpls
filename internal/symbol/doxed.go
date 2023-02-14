package symbol

import (
	"fmt"
	"log"
	"strings"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/VKCOM/php-parser/pkg/position"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/typer"
)

type DocFilter func(phpdoxer.Node) bool

func FilterDocKind(kind phpdoxer.NodeKind) DocFilter {
	return func(n phpdoxer.Node) bool {
		return n.Kind() == kind
	}
}

type Doxed struct {
	node ir.Node

	docNodeCache []phpdoxer.Node
}

func NewDoxed(node ir.Node) *Doxed {
	return &Doxed{node: node}
}

func (d *Doxed) Docs() []phpdoxer.Node {
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

func (d *Doxed) FindDoc(filters ...DocFilter) phpdoxer.Node {
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

func (d *Doxed) FindAllDocs(filters ...DocFilter) (results []phpdoxer.Node) {
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

func (d *Doxed) RawDocs() string {
	return strings.Join(typer.NodeComments(d.node), "\n")
}

// DocClassNormalize applies any context we have about the code (which phpdoxer
// does not have).
//
// Normalizations:
// 1. "static" -> current class
// 2. "$this" -> current class
// 3. A class in all CAPS -> qualified class
// 4. Precedence -> recursively unpacked
// 5. Union -> recursively unpacked
// 6. Intersection -> recursively unpacked
//
// Cases 4, 5 and 6 can result in multiple classes, so a slice is returned.
func DocClassNormalize(
	fqnt *fqn.Traverser,
	currFqn *fqn.FQN,
	currPos *position.Position,
	doc phpdoxer.Type,
) []*phpdoxer.TypeClassLike {
	switch typed := doc.(type) {
	case *phpdoxer.TypeClassLike:
		switch typed.Name {
		case "self", "static", "$this":
			return []*phpdoxer.TypeClassLike{{
				Name:           currFqn.String(),
				FullyQualified: true,
			}}
		}

		return []*phpdoxer.TypeClassLike{{
			Name:           fqnt.ResultFor(&ir.Name{Position: currPos, Value: typed.Name}).String(),
			FullyQualified: true,
		}}
	case *phpdoxer.TypeConstant:
		if typed.Class != nil {
			return nil
		}

		// In case the user created a class in all CAPS, we catch here that
		// it is not a constant, but a class.
		res := fqnt.ResultFor(&ir.Name{Position: currPos, Value: typed.Const})
		if res != nil {
			return []*phpdoxer.TypeClassLike{{Name: res.String(), FullyQualified: true}}
		}

	case *phpdoxer.TypePrecedence:
		return DocClassNormalize(fqnt, currFqn, currPos, typed.Type)
	case *phpdoxer.TypeUnion:
		var res []*phpdoxer.TypeClassLike
		res = append(res, DocClassNormalize(fqnt, currFqn, currPos, typed.Left)...)
		res = append(res, DocClassNormalize(fqnt, currFqn, currPos, typed.Right)...)
		return res
	case *phpdoxer.TypeIntersection:
		var res []*phpdoxer.TypeClassLike
		res = append(res, DocClassNormalize(fqnt, currFqn, currPos, typed.Left)...)
		res = append(res, DocClassNormalize(fqnt, currFqn, currPos, typed.Right)...)
		return res
	}

	return nil
}
