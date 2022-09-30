package project

import (
	"errors"
	"fmt"
	"log"

	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/common"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/internal/project/definition/providers"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/position"
)

var (
	definitionProviders = []DefinitionProvider{
		providers.NewThis(),     // $this
		providers.NewFunction(), // explode()
		providers.NewVariable(), // $a
		providers.NewUse(),      // use Foo, use Foo\Bar\FooBar as Foo
		providers.NewConstant(), // FOO, BAR, FOOBAR
		providers.NewName(),     // new Class, Class::
		providers.NewProperty(), // $this->foo, $foo->foo->bar
		providers.NewMethod(),   // $this->test(), $foo->foo->test()
		providers.NewStatic(),   // Foo::bar(), Foo::bar()->baz() // TODO: add tests.
	}

	ErrNoDefinitionFound = errors.New(
		"No definition found for symbol at given position",
	)
)

type DefinitionProvider interface {
	CanDefine(ctx context.Context, kind ir.NodeKind) bool
	Define(ctx context.Context) ([]*definition.Definition, error)
}

func (p *Project) Definition(pos *position.Position) ([]*position.Position, error) {
	ctx, err := context.New(pos)
	if err != nil {
		return nil, fmt.Errorf("Could not create definition context: %w", err)
	}

	for advanced := true; advanced; advanced = ctx.Advance() {
		kind := ir.GetNodeKind(ctx.Current())

		for _, provider := range definitionProviders {
			if provider.CanDefine(ctx, kind) {
				what.Happens("Defining using provider %T", provider)

				defs, err := provider.Define(ctx)
				if err != nil {
					log.Println(err)
					return nil, ErrNoDefinitionFound
				}

				return common.Map(defs, func(def *definition.Definition) *position.Position {
					pos, err := defPosition(def)
					if err != nil {
						log.Println(err)
						return nil
					}

					return pos
				}), nil
			}
		}
	}

	log.Println("no definition provider registered for the given position")
	return nil, ErrNoDefinitionFound
}

func defPosition(def *definition.Definition) (*position.Position, error) {
	content, err := wrkspc.FromContainer().ContentOf(def.Path)
	if err != nil {
		log.Println(err)
		return nil, ErrNoDefinitionFound
	}

	pos := def.Node.Position()
	_, col := position.PosToLoc(content, uint(pos.StartPos))

	return &position.Position{
		Row:  uint(pos.StartLine),
		Col:  col,
		Path: def.Path,
	}, nil
}
