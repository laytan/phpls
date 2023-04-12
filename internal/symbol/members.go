package symbol

import (
	"fmt"
	"log"

	"github.com/laytan/php-parser/pkg/ast"
)

type Member interface {
	NamedModified
	Vertex() ast.Vertex
	member()
}

func (p *Property) member() {}
func (m *Method) member()   {}

func (c *ClassLike) FindMember(filters ...FilterFunc[Member]) Member {
	res := c.FindMembers(true, filters...)
	if len(res) == 0 {
		return nil
	}

	return res[0]
}

// TODO: also constants.
func (c *ClassLike) FindMembers(shortCircuit bool, filters ...FilterFunc[Member]) (res []Member) {
	miter := c.MethodsIter()
MethodsIter:
	for m, done, err := miter(); !done; m, done, err = miter() {
		if err != nil {
			log.Println(fmt.Errorf("[symbol.ClassLike.FindMethod]: %w", err))
			continue
		}

		for _, filter := range filters {
			if !filter(m) {
				continue MethodsIter
			}
		}

		res = append(res, m)
		if shortCircuit {
			return res
		}
	}

	piter := c.PropertiesIter()
PropertiesIter:
	for p, done, err := piter(); !done; p, done, err = piter() {
		if err != nil {
			log.Println(fmt.Errorf("[symbol.ClassLike.FindProperty]: %w", err))
			continue
		}

		for _, filter := range filters {
			if !filter(p) {
				continue PropertiesIter
			}
		}

		res = append(res, p)
		if shortCircuit {
			return res
		}
	}

	return res
}

func (c *ClassLike) FindMemberInherit(filters ...FilterFunc[Member]) Member {
	res := c.FindMembers(true, filters...)
	if len(res) == 0 {
		return nil
	}

	return res[0]
}

func (c *ClassLike) FindMembersInherit(
	shortCircuit bool,
	filters ...FilterFunc[Member],
) (res []Member) {
	iter := c.InheritsIter()
	for cls, done, err := iter(); !done; cls, done, err = iter() {
		if err != nil {
			log.Printf(
				"Iterating inherited classes of %q failed while finding members: %v",
				c.Name(),
				err,
			)
			continue
		}

		res = append(res, cls.FindMembers(shortCircuit, filters...)...)
		if shortCircuit && len(res) > 0 {
			return res
		}
	}

	return res
}
