package providers

import (
	"errors"
	"fmt"
	"log"

	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project/definition"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/resolvequeue"
	"github.com/laytan/elephp/pkg/stack"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
)

// PropertyProvider resolves the definition of a property accessed like $a->property.
// Where $a can also be $this, $a->foo->bar()->property etc.
type PropertyProvider struct{}

func NewProperty() *PropertyProvider {
	return &PropertyProvider{}
}

func (p *PropertyProvider) CanDefine(ctx context.Context, kind ir.NodeKind) bool {
	return kind == ir.KindPropertyFetchExpr
}

// Recurse all the way down the .Variable, until we get to a *ir.SimpleVar,
// get the type of that variable,
// go back up resolving the types as we go.
func (p *PropertyProvider) Define(ctx context.Context) (*definition.Definition, error) {
	n := ctx.Current().(*ir.PropertyFetchExpr)

	properties := stack.New[string]()
	def, privacy, err := p.down(ctx, properties, n)
	if err != nil {
		return nil, err
	}

	def, err = p.up(ctx, def, privacy, properties)
	if err != nil {
		if errors.Is(err, definition.ErrNoDefinitionFound) {
			return nil, err
		}

		log.Println(err)
		return nil, definition.ErrNoDefinitionFound
	}

	return def, nil
}

// Keeps going down the call until a variable is hit, keeping track of where
// we have been in the properties stack.
//
// Will end up with the root ($this in $this->a->b, $foo in $foo->bar->baz)'s
// type definition and the relation of the context with it in terms of access (privacy).
func (p *PropertyProvider) down(
	ctx context.Context,
	properties *stack.Stack[string],
	fetch *ir.PropertyFetchExpr,
) (*definition.Definition, phprivacy.Privacy, error) {
	name, ok := fetch.Property.(*ir.Identifier)
	if !ok {
		return nil, 0, fmt.Errorf(
			definition.ErrUnexpectedNodeFmt,
			fetch.Property,
			"*ir.Identifier",
		)
	}

	properties.Push(name.Value)

	// TODO: support methods here (get the return type, keep going).
	switch variable := fetch.Variable.(type) {
	case *ir.SimpleVar:
		// Base case, get variable type.
		return definition.VariableType(ctx, variable)

	case *ir.PropertyFetchExpr:
		// Recursively call this.
		return p.down(ctx, properties, variable)

	default:
		return nil, 0, fmt.Errorf(definition.ErrUnexpectedNodeFmt, fetch.Variable, "*ir.SimpleVar, *ir.PropertyFetchExpr or *ir.MethodCallExpr")
	}
}

// Walks back up the properties stack, resolving the types, returning the last type.
func (p *PropertyProvider) up(
	ctx context.Context,
	start *definition.Definition,
	privacy phprivacy.Privacy,
	properties *stack.Stack[string],
) (*definition.Definition, error) {
	what.Happens(
		"Walking properties, starting from %s->%s\n",
		start.Node.Identifier(),
		properties.Peek(),
	)

	currentDef := start
	currentRoot := ctx.Root()

	var resultProp *ir.PropertyListStmt
	var resultPath string
	for prop := properties.Pop(); prop != ""; prop = properties.Pop() {
		// walk resolve queue
		err := walkResolveQueue(
			ctx,
			currentRoot,
			currentDef.Node,
			func(wc *walkContext) (bool, error) {
				what.Happens("Checking %s for property %s\n", wc.QueueNode.FQN.String(), prop)

				propTraverser := traversers.NewProperty(
					prop,
					wc.QueueNode.FQN.Name(),
					privacy,
				)
				wc.IR.Walk(propTraverser)

				if propTraverser.Property == nil {
					what.Happens(
						"Could not find property %s in %s\n",
						prop,
						wc.QueueNode.FQN.String(),
					)

					if !wc.HasMore {
						what.Happens("No more to go")
						return true, definition.ErrNoDefinitionFound
					}

					what.Happens("more to go")
					return false, nil
				}

				resultProp = propTraverser.Property
				resultPath = wc.CurrDef.Path

				// get property type (classLike)
				propType := ctx.Typer().Property(currentRoot, propTraverser.Property)
				if propType == nil {
					return true, nil
				}

				if clsType, ok := propType.(*phpdoxer.TypeClassLike); ok {
					def, ok := definition.FindFullyQualified(
						currentRoot,
						ctx.Index(),
						clsType.Name,
						symbol.ClassLikeScopes...)
					if !ok {
						return false, nil
					}

					what.Happens(
						"Found class-like %s for property %s\n",
						def.Node.Identifier(),
						prop,
					)

					newRoot, err := ctx.Workspace().IROf(def.Path)
					if err != nil {
						return false, fmt.Errorf(
							"Walking properties: unable to get file content/nodes of %s: %w",
							def.Path,
							err,
						)
					}

					currentDef = def
					currentRoot = newRoot
				}

				return true, nil
			},
		)
		if err != nil {
			resultProp = nil
			resultPath = ""
			break
		}
	}

	if resultProp == nil {
		return nil, definition.ErrNoDefinitionFound
	}

	return &definition.Definition{Path: resultPath, Node: symbol.New(resultProp)}, nil
}

type walkContext struct {
	QueueNode *resolvequeue.Node
	CurrDef   *definition.Definition
	Privacy   phprivacy.Privacy
	IR        *ir.Root
	HasMore   bool
}

func walkResolveQueue(
	ctx context.Context,
	root *ir.Root,
	start symbol.Symbol,
	walker func(*walkContext) (bool, error),
) error {
	fqn := definition.FullyQualify(root, start.Identifier())
	resolveQueue := resolvequeue.New(
		func(n *resolvequeue.Node) (*ir.Root, error) {
			def, err := ctx.Index().Find(n.FQN.String(), n.Kind)
			if err != nil {
				if !errors.Is(err, index.ErrNotFound) {
					log.Println(err)
				}

				return nil, definition.ErrNoDefinitionFound
			}

			return ctx.Workspace().IROf(def.Path)
		},
		&resolvequeue.Node{FQN: fqn, Kind: start.NodeKind()},
	)

	isCurr := true
	for res := resolveQueue.Queue.Dequeue(); res != nil; func() {
		res = resolveQueue.Queue.Dequeue()
		isCurr = false
	}() {
		def, err := ctx.Index().Find(res.FQN.String(), res.Kind)
		if err != nil {
			if !errors.Is(err, index.ErrNotFound) {
				log.Println(err)
			}

			return definition.ErrNoDefinitionFound
		}

		// NOTE: this is only correct when resolvequeue is called for a symbol
		// NOTE: inside of the class. Not for $variable->method() for example.
		var privacy phprivacy.Privacy
		switch def.Symbol.NodeKind() {
		case ir.KindClassStmt:
			// If first index (source file) search for any privacy,
			// if not, search for protected and public privacy.
			privacy = phprivacy.PrivacyProtected

			if isCurr {
				privacy = phprivacy.PrivacyPrivate
			}
		default:
			// Traits and interface members are available everywhere.
			privacy = phprivacy.PrivacyPrivate
		}

		root, err := ctx.Workspace().IROf(def.Path)
		if err != nil {
			return fmt.Errorf(def.Path, "Error parsing ast for %s: %w", err)
		}

		done, err := walker(
			&walkContext{
				res,
				&definition.Definition{Path: def.Path, Node: def.Symbol},
				privacy,
				root,
				resolveQueue.Queue.Peek() != nil,
			},
		)
		if err != nil {
			return err
		}

		if done {
			break
		}
	}

	return nil
}
