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
	"github.com/laytan/elephp/pkg/typer"
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
	id, ok := n.Property.(*ir.Identifier)
	if !ok {
		log.Println(fmt.Errorf(definition.ErrUnexpectedNodeFmt, n.Property, "*ir.Identifier"))
		return nil, definition.ErrNoDefinitionFound
	}

	def, privacy, err := p.down(ctx, properties, id.Value, n.Variable)
	if err != nil {
		log.Println(err)
		return nil, definition.ErrNoDefinitionFound
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
	currentSymbol string,
	currentVar ir.Node,
) (*definition.Definition, phprivacy.Privacy, error) {
	properties.Push(currentSymbol)

	switch variable := currentVar.(type) {
	case *ir.SimpleVar:
		// Base case, get variable type.
		return definition.VariableType(ctx, variable)

	case *ir.PropertyFetchExpr:
		// Recursively call this.
		id, ok := variable.Property.(*ir.Identifier)
		if !ok {
			return nil, 0, fmt.Errorf(definition.ErrUnexpectedNodeFmt, variable.Property, "*ir.Identifier")
		}

		return p.down(ctx, properties, id.Value, variable.Variable)

	case *ir.MethodCallExpr:
		// Recursively call this.
		id, ok := variable.Method.(*ir.Identifier)
		if !ok {
			return nil, 0, fmt.Errorf(definition.ErrUnexpectedNodeFmt, variable.Method, "*ir.Identifier")
		}

		return p.down(ctx, properties, id.Value+"()", variable.Variable)

	default:
		return nil, 0, fmt.Errorf(definition.ErrUnexpectedNodeFmt, currentVar, "*ir.SimpleVar, *ir.PropertyFetchExpr or *ir.MethodCallExpr")
	}
}

// Walks back up the properties stack, resolving the types, returning the last type.
func (p *PropertyProvider) up(
	ctx context.Context,
	start *definition.Definition,
	privacy phprivacy.Privacy,
	symbols *stack.Stack[string],
) (*definition.Definition, error) {
	what.Happens(
		"Walking properties, starting from %s->%s\n",
		start.Node.Identifier(),
		symbols.Peek(),
	)

	currentDef := start
	currentRoot := ctx.Root()

	var resultSymbol ir.Node
	var resultPath string
	for prop := symbols.Pop(); prop != ""; prop = symbols.Pop() {
		// walk resolve queue
		err := walkResolveQueue(
			ctx,
			currentRoot,
			currentDef.Node,
			func(wc *walkContext) (bool, error) {
				what.Happens("Checking %s for symbol %s\n", wc.FQN, prop)
				root, err := ctx.Workspace().IROf(wc.Curr.Path)
				if err != nil {
					return true, err
				}

				var result ir.Node
				switch prop[len(prop)-2:] {
				case "()":
					traverser := traversers.NewMethod(prop[:len(prop)-2], wc.FQN.Name(), privacy)
					root.Walk(traverser)
					if traverser.Method != nil {
						result = traverser.Method
					}

				default:
					traverser := traversers.NewProperty(prop, wc.FQN.Name(), privacy)
					root.Walk(traverser)
					if traverser.Property != nil {
						result = traverser.Property
					}
				}

				if result == nil {
					what.Happens(
						"Could not find symbol %s in %s\n",
						prop,
						wc.FQN,
					)

					if wc.IsLast {
						return true, definition.ErrNoDefinitionFound
					}

					return false, nil
				}

				// At this point we have found the property in this class.
				what.Happens(
					"Found symbol definition for %s in %s",
					prop,
					wc.Curr.Node.Identifier(),
				)

				resultSymbol = result
				resultPath = wc.Curr.Path

				var symType phpdoxer.Type
				switch symbol := result.(type) {
				case *ir.PropertyListStmt:
					// get property type.
					propType := ctx.Typer().Property(currentRoot, symbol)
					if propType == nil {
						what.Happens("No type found for property %s in %s", prop, wc.FQN)
						return true, nil
					}

					symType = propType

				case *ir.ClassMethodStmt:
					// Get method return type.
					methType := ctx.Typer().Returns(currentRoot, symbol)
					if methType == nil {
						what.Happens("No type found for method %s in %s", prop, wc.FQN)
						return true, nil
					}

					symType = methType

				default:
					return true, fmt.Errorf(definition.ErrUnexpectedNodeFmt, result, "*ir.PropertyListStmt or *ir.ClassMethodStmt")
				}

				if symType.Kind() != phpdoxer.KindClassLike {
					return true, fmt.Errorf(
						definition.ErrUnexpectedNodeFmt,
						symType,
						"*phpdoxer.TypeClassLike",
					)
				}

				clsType := symType.(*phpdoxer.TypeClassLike)
				def, ok := definition.FindFullyQualified(
					currentRoot,
					ctx.Index(),
					clsType.Name,
					symbol.ClassLikeScopes...)
				if !ok {
					return false, nil
				}

				what.Happens(
					"Found class-like %s for symbol %s\n",
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

				return true, nil
			},
		)
		if err != nil {
			resultSymbol = nil
			resultPath = ""
			break
		}
	}

	if resultSymbol == nil {
		return nil, definition.ErrNoDefinitionFound
	}

	return &definition.Definition{Path: resultPath, Node: symbol.New(resultSymbol)}, nil
}

// walkContext is the context of a current iteration in the walk of the resolve queue.
type walkContext struct {
	// The FQN of the current class.
	FQN *typer.FQN

	// The definition of the current class.
	Curr *definition.Definition

	// Whether there are more classes in the resolve queue.
	IsLast bool
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

	for res := resolveQueue.Queue.Dequeue(); res != nil; res = resolveQueue.Queue.Dequeue() {
		def, err := ctx.Index().Find(res.FQN.String(), res.Kind)
		if err != nil {
			if !errors.Is(err, index.ErrNotFound) {
				log.Println(err)
			}

			return definition.ErrNoDefinitionFound
		}

		done, err := walker(
			&walkContext{
				FQN:    res.FQN,
				Curr:   &definition.Definition{Node: def.Symbol, Path: def.Path},
				IsLast: resolveQueue.Queue.Peek() == nil,
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
