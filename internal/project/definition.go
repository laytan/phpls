package project

import (
	"errors"
	"fmt"
	"log"

	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/internal/project/definition/providers"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/position"
)

var (
	definitionProviders = []DefinitionProvider{
		providers.NewComments(),      // Class names in comment tags.
		providers.NewThis(),          // $this
		providers.NewFunction(),      // explode()
		providers.NewVariable(),      // $a
		providers.NewUse(),           // use Foo, use Foo\Bar\FooBar as Foo
		providers.NewConstant(),      // FOO, BAR, FOOBAR
		providers.NewName(),          // new Class, Class::
		providers.NewProperty(),      // $this->foo, $foo->foo->bar
		providers.NewMethod(),        // $this->test(), $foo->foo->test()
		providers.NewClassConstant(), // Foo::BAR, self::BAR, $this::FOO
		providers.NewStatic(),        // Foo::bar(), Foo::bar()->baz() // TODO: add tests.
	}

	ErrNoDefinitionFound = errors.New(
		"No definition found for symbol at given position",
	)
)

type DefinitionProvider interface {
	CanDefine(ctx *context.Ctx, kind ir.NodeKind) bool
	Define(ctx *context.Ctx) ([]*definition.Definition, error)
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

				return functional.MapFilter(defs, defPosition), nil
			}
		}
	}

	log.Println("no definition provider registered for the given position")
	return nil, ErrNoDefinitionFound
}

func defPosition(def *definition.Definition) *position.Position {
	content := wrkspc.FromContainer().FContentOf(def.Path)
	return position.FromIRPosition(def.Path, content, def.Position.StartPos)
}
