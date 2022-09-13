package definition

import (
	"errors"
	"fmt"
	"log"

	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/context"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/resolvequeue"
	"github.com/laytan/elephp/pkg/stack"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
)

func WalkToProperty(ctx context.Context, start *ir.PropertyFetchExpr) (*Definition, error) {
	id, ok := start.Property.(*ir.Identifier)
	if !ok {
		log.Println(fmt.Errorf(ErrUnexpectedNodeFmt, start.Property, "*ir.Identifier"))
		return nil, ErrNoDefinitionFound
	}

	return walkToClassSymbol(ctx, id.Value, start.Variable)
}

func WalkToMethod(ctx context.Context, start *ir.MethodCallExpr) (*Definition, error) {
	id, ok := start.Method.(*ir.Identifier)
	if !ok {
		log.Println(fmt.Errorf(ErrUnexpectedNodeFmt, start.Method, "*ir.Identifier"))
		return nil, ErrNoDefinitionFound
	}

	return walkToClassSymbol(ctx, id.Value+"()", start.Variable)
}

func walkToClassSymbol(ctx context.Context, start string, startVar ir.Node) (*Definition, error) {
	symbols := stack.New[string]()
	def, privacy, err := down(ctx, symbols, start, startVar)
	if err != nil {
		log.Println(err)
		return nil, ErrNoDefinitionFound
	}

	def, err = up(ctx, def, privacy, symbols)
	if err != nil {
		if errors.Is(err, ErrNoDefinitionFound) {
			return nil, err
		}

		log.Println(err)
		return nil, ErrNoDefinitionFound
	}

	return def, nil
}

// TODO: combine into one function, and give proper name.

// Keeps going Down the call until a variable is hit, keeping track of where
// we have been in the properties stack.
//
// Will end up with the root ($this in $this->a->b, $foo in $foo->bar->baz)'s
// type definition and the relation of the context with it in terms of access (privacy).
func down(
	ctx context.Context,
	properties *stack.Stack[string],
	currentSymbol string,
	currentVar ir.Node,
) (*Definition, phprivacy.Privacy, error) {
	properties.Push(currentSymbol)

	switch variable := currentVar.(type) {
	case *ir.SimpleVar:
		// Base case, get variable type.
		return VariableType(ctx, variable)

	case *ir.PropertyFetchExpr:
		// Recursively call this.
		id, ok := variable.Property.(*ir.Identifier)
		if !ok {
			return nil, 0, fmt.Errorf(ErrUnexpectedNodeFmt, variable.Property, "*ir.Identifier")
		}

		return down(ctx, properties, id.Value, variable.Variable)

	case *ir.MethodCallExpr:
		// Recursively call this.
		id, ok := variable.Method.(*ir.Identifier)
		if !ok {
			return nil, 0, fmt.Errorf(ErrUnexpectedNodeFmt, variable.Method, "*ir.Identifier")
		}

		return down(ctx, properties, id.Value+"()", variable.Variable)

	default:
		return nil, 0, fmt.Errorf(ErrUnexpectedNodeFmt, currentVar, "*ir.SimpleVar, *ir.PropertyFetchExpr or *ir.MethodCallExpr")
	}
}

// Walks back up the properties stack, resolving the types, returning the last type.
func up(
	ctx context.Context,
	start *Definition,
	privacy phprivacy.Privacy,
	symbols *stack.Stack[string],
) (*Definition, error) {
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
						return true, ErrNoDefinitionFound
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
					return true, fmt.Errorf(ErrUnexpectedNodeFmt, result, "*ir.PropertyListStmt or *ir.ClassMethodStmt")
				}

				if symType.Kind() != phpdoxer.KindClassLike {
					return true, fmt.Errorf(
						ErrUnexpectedNodeFmt,
						symType,
						"*phpdoxer.TypeClassLike",
					)
				}

				clsType := symType.(*phpdoxer.TypeClassLike)
				def, ok := FindFullyQualified(
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
		return nil, ErrNoDefinitionFound
	}

	return &Definition{Path: resultPath, Node: symbol.New(resultSymbol)}, nil
}

// walkContext is the context of a current iteration in the walk of the resolve queue.
type walkContext struct {
	// The FQN of the current class.
	FQN *typer.FQN

	// The definition of the current class.
	Curr *Definition

	// Whether there are more classes in the resolve queue.
	IsLast bool
}

func walkResolveQueue(
	ctx context.Context,
	root *ir.Root,
	start symbol.Symbol,
	walker func(*walkContext) (bool, error),
) error {
	fqn := FullyQualify(root, start.Identifier())
	resolveQueue := resolvequeue.New(
		func(n *resolvequeue.Node) (*ir.Root, error) {
			def, err := ctx.Index().Find(n.FQN.String(), n.Kind)
			if err != nil {
				if !errors.Is(err, index.ErrNotFound) {
					log.Println(err)
				}

				return nil, ErrNoDefinitionFound
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

			return ErrNoDefinitionFound
		}

		done, err := walker(
			&walkContext{
				FQN:    res.FQN,
				Curr:   &Definition{Node: def.Symbol, Path: def.Path},
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
