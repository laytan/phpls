package project

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/resolvequeue"
	"github.com/laytan/elephp/pkg/symbol"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/elephp/pkg/typer"
)

var (
	returnTypeRegex = regexp.MustCompile(`@return ([\w\\]+)`)
	varRegex        = regexp.MustCompile(`@var ([\w\\]+)`)
)

// Resolves the fully qualified name for the given name node.
//
// This resolves, use statements, aliassed use statements are resolved to the
// non-aliassed version.
func (p *Project) FQN(root *ir.Root, name *ir.Name) *typer.FQN {
	if name.IsFullyQualified() {
		return typer.NewFQN(name.Value)
	}

	traverser := typer.NewFQNTraverser()
	root.Walk(traverser)

	return traverser.ResultFor(name)
}

func (p *Project) assignment(scope ir.Node, variable *ir.SimpleVar) *ir.SimpleVar {
	traverser := traversers.NewAssignment(variable)
	scope.Walk(traverser)

	return traverser.Assignment
}

func (p *Project) globalAssignment(root *ir.Root, globalVar *ir.SimpleVar) *ir.SimpleVar {
	// First search the current file for the assignment.
	traverser := traversers.NewGlobalAssignment(globalVar)
	root.Walk(traverser)

	// TODO: search the whole project if the global is not assigned here.

	return traverser.Assignment
}

func (p *Project) function(
	scope ir.Node,
	call *ir.FunctionCallExpr,
) (*symbol.FunctionStmtSymbol, string) {
	traverser, err := traversers.NewFunction(call)
	if err != nil {
		return nil, ""
	}

	scope.Walk(traverser)
	if traverser.Function != nil {
		return symbol.NewFunction(traverser.Function), ""
	}

	// No definition found locally, searching globally.
	name, ok := call.Function.(*ir.Name)
	if !ok {
		return nil, ""
	}

	results := p.symbolTrie.SearchExact(name.Value)

	if len(results) == 0 {
		return nil, ""
	}

	for _, res := range results {
		if function, ok := res.Symbol.(*symbol.FunctionStmtSymbol); ok {
			return function, res.Path
		}
	}

	return nil, ""
}

func (p *Project) classLike(
	root *ir.Root,
	name *ir.Name,
) (*symbol.ClassLikeStmtSymbol, string) {
	fqn := p.FQN(root, name)

	results := p.symbolTrie.SearchExact(fqn.Name())

	if len(results) == 0 {
		return nil, ""
	}

	for _, res := range results {
		if res.Namespace != fqn.Namespace() {
			continue
		}

		if symbol, ok := res.Symbol.(*symbol.ClassLikeStmtSymbol); ok {
			return symbol, res.Path
		}
	}

	return nil, ""
}

type rootRetriever struct {
	project *Project
}

func (r *rootRetriever) RetrieveRoot(n *resolvequeue.Node) (*ir.Root, error) {
	file := r.project.FindFileInTrie(n.FQN, n.Kind)
	if file == nil {
		return nil, fmt.Errorf("No file indexed for FQN %s", n.FQN)
	}

	ast := r.project.ParseFileCached(file)
	if ast == nil {
		return nil, fmt.Errorf("Error parsing file for FQN %s", n.FQN)
	}

	return ast, nil
}

func (p *Project) method(
	root *ir.Root,
	classLikeScope ir.Node,
	scope ir.Node,
	method *ir.MethodCallExpr,
) (*ir.ClassMethodStmt, string, error) {
	objectVar, ok := method.Variable.(*ir.SimpleVar)
	if !ok {
		return nil, "", fmt.Errorf(
			"Method definition: unexpected variable node of type %T",
			method.Variable,
		)
	}

	methodName, ok := method.Method.(*ir.Identifier)
	if !ok {
		return nil, "", fmt.Errorf(
			"Method definition: unexpected method node of type %T",
			method.Method,
		)
	}

	classScope, privacy, err := p.variableType(root, classLikeScope, scope, objectVar)
	if err != nil {
		return nil, "", fmt.Errorf("Error fetching variable type: %w", err)
	}

	if classScope == nil {
		return nil, "", nil
	}

	file := p.GetFile(classScope.Path)
	classRoot := p.ParseFileCached(file)

	var result *ir.ClassMethodStmt
	var resultPath string
	err = p.walkResolveQueue(
		classRoot,
		classScope.Symbol,
		func(wc *walkContext) (bool, error) {
			traverser := traversers.NewMethod(methodName.Value, wc.QueueNode.FQN.Name(), privacy)
			wc.IR.Walk(traverser)

			if traverser.Method != nil {
				result = traverser.Method
				resultPath = wc.TrieNode.Path

				return true, nil
			}

			return false, nil
		},
	)
	if err != nil {
		return nil, "", fmt.Errorf(
			"Error retrieving method definition for %s->%s: %w",
			objectVar.Name,
			methodName.Value,
			err,
		)
	}

	return result, resultPath, nil
}

func (p *Project) property(
	root *ir.Root,
	classLikeScope ir.Node,
	scope ir.Node,
	stmt *ir.PropertyFetchExpr,
) (*ir.PropertyListStmt, string, error) {
	objectVar, ok := stmt.Variable.(*ir.SimpleVar)
	if !ok {
		return nil, "", fmt.Errorf(
			"Property definition: unexpected variable node of type %T",
			stmt.Variable,
		)
	}

	propName, ok := stmt.Property.(*ir.Identifier)
	if !ok {
		return nil, "", fmt.Errorf(
			"Property definition: unexpected property node of type %T",
			stmt.Property,
		)
	}

	classScope, privacy, err := p.variableType(root, classLikeScope, scope, objectVar)
	if err != nil {
		return nil, "", fmt.Errorf("Error fetching variable type: %w", err)
	}

	if classScope == nil {
		return nil, "", nil
	}

	file := p.GetFile(classScope.Path)
	classRoot := p.ParseFileCached(file)

	var result *ir.PropertyListStmt
	var resultPath string
	err = p.walkResolveQueue(classRoot, classScope.Symbol, func(wc *walkContext) (bool, error) {
		traverser := traversers.NewProperty(propName.Value, wc.QueueNode.FQN.Name(), privacy)
		wc.IR.Walk(traverser)

		if traverser.Property != nil {
			result = traverser.Property
			resultPath = wc.TrieNode.Path

			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, "", fmt.Errorf("Property definition: %w", err)
	}

	return result, resultPath, nil
}

type walkContext struct {
	QueueNode *resolvequeue.Node
	TrieNode  *traversers.TrieNode
	File      *File
	Privacy   phprivacy.Privacy
	IR        *ir.Root
}

func (p *Project) walkResolveQueue(
	root *ir.Root,
	classLikeScope symbol.Symbol,
	walker func(*walkContext) (bool, error),
) error {
	fqn := p.FQN(root, &ir.Name{Value: classLikeScope.Identifier()})
	resolveQueue := resolvequeue.New(
		&rootRetriever{project: p},
		&resolvequeue.Node{FQN: fqn, Kind: classLikeScope.NodeKind()},
	)

	isCurr := true
	for res := resolveQueue.Queue.Dequeue(); res != nil; func() {
		res = resolveQueue.Queue.Dequeue()
		isCurr = false
	}() {
		file, symbol := p.FindFileAndSymbolInTrie(res.FQN, res.Kind)
		if file == nil {
			return fmt.Errorf("Can't get file for FQN: %s", res.FQN.String())
		}

		// NOTE: this is only correct when resolvequeue is called for a symbol
		// NOTE: inside of the class. Not for $variable->method() for example.
		var privacy phprivacy.Privacy
		switch symbol.Symbol.NodeKind() {
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

		ast := p.ParseFileCached(file)
		if ast == nil {
			return fmt.Errorf("Error parsing ast for %s", file.path)
		}

		done, err := walker(&walkContext{res, symbol, file, privacy, ast})
		if err != nil {
			return err
		}

		if done {
			break
		}
	}

	return nil
}

// TODO: union types.
func (p *Project) variableType(
	root *ir.Root,
	classScope ir.Node,
	scope ir.Node,
	variable *ir.SimpleVar,
) (*traversers.TrieNode, phprivacy.Privacy, error) {
	switch variable.Name {
	case "this", "self", "static":
		fqn := p.FQN(root, &ir.Name{Value: symbol.GetIdentifier(classScope)})
		node := p.FindNodeInTrieMultiKinds(fqn, symbol.ClassLikeScopes)
		if node == nil {
			return nil, 0, fmt.Errorf(
				"Unable to get type of %s",
				variable.Name,
			)
		}

		return node, phprivacy.PrivacyPrivate, nil

	case "parent":
		parent, err := p.parentOf(root, classScope)
		if err != nil {
			return nil, 0, fmt.Errorf("Could not resolve parent: %w", err)
		}

		return parent, phprivacy.PrivacyProtected, nil

	default:
		// Get the assignment of the variable.
		traverser := traversers.NewAssignment(variable)
		scope.Walk(traverser)
		if traverser.Assignment == nil || traverser.Scope == nil {
			return nil, 0, fmt.Errorf("Could not find assignment for variable %s", variable.Name)
		}

		// If scope is a parameter, see if the scope(function/method, passed into this func) has doc with @param,
		// If not, see if the param has type hints,
		// If not, we don't know the type.
		switch varScope := traverser.Scope.(type) {
		case *ir.Parameter:
			switch scope.(type) {
			case *ir.FunctionStmt, *ir.ClassMethodStmt:
				paramType := p.typer.Param(root, scope, varScope)

				if typedRetType, ok := paramType.(*phpdoxer.TypeClassLike); ok {
					retSym := p.findClassLikeSymbol(typedRetType)
					return retSym, phprivacy.PrivacyPublic, nil
				}

				return nil, 0, nil

			default:
				return nil, 0, fmt.Errorf("Given variable %s is a parameter, but scope is %T, expected *ir.FunctionStmt or *ir.ClassMethodStmt", variable.Name, scope)
			}

		case *ir.GlobalStmt:
			panic("Unimplemented")

		case *ir.Assign:
			// If scope is assign, check traverser.assignment for a phpdoc.
			// If it has a phpdoc, with @var, return the symbol for that type.
			varType := p.typer.Variable(root, traverser.Assignment, scope)

			if clsLike, ok := varType.(*phpdoxer.TypeClassLike); ok {
				assignment := p.findClassLikeSymbol(clsLike)
				if assignment != nil {
					// TODO: privacy will be wrong in some cases.
					return assignment, phprivacy.PrivacyPrivate, nil
				}
			}

			switch exprNode := varScope.Expr.(type) {
			// Do this recursively, with .Expr:
			case *ir.CloneExpr, *ir.ParenExpr, *ir.ErrorSuppressExpr:
				// Same but its a Node:
			case *ir.ReferenceExpr:
				// Functions/methods to get return type from:
			case *ir.ClosureExpr, *ir.MethodCallExpr, *ir.ArrowFunctionExpr, *ir.FunctionCallExpr, *ir.StaticCallExpr, *ir.NullsafeMethodCallExpr:
				// Return left or right type (union):
			case *ir.CoalesceExpr, *ir.TernaryExpr:
				// Return the type being instantiated:
			case *ir.NewExpr:
				if className, ok := exprNode.Class.(*ir.Name); ok {
					fqn := p.FQN(root, className)
					node := p.FindNodeInTrie(fqn, ir.KindClassStmt)
					if node == nil {
						return nil, 0, fmt.Errorf("Could not resolve FQN of name %s", className.Value)
					}

					return node, phprivacy.PrivacyPublic, nil
				}

				return nil, 0, fmt.Errorf("Expected new node's class node to be of type *ir.Name, got: %T", exprNode.Class)

				// Fetch properties/variables:
			case *ir.ConstFetchExpr, *ir.PropertyFetchExpr, *ir.ClassConstFetchExpr, *ir.NullsafePropertyFetchExpr, *ir.StaticPropertyFetchExpr:
				// Return the type being casted:
			case *ir.TypeCastExpr:
			// Don't know how:
			case *ir.MatchExpr, *ir.ArrayDimFetchExpr, *ir.AnonClassExpr:
			default:
				return nil, 0, fmt.Errorf("Unable to find type of variable from usage/expression, got node type %T", varScope.Expr)

			}
		}

		return nil, 0, nil
	}
}

func (p *Project) parentOf(root *ir.Root, classScope ir.Node) (*traversers.TrieNode, error) {
	classStmt, ok := classScope.(*ir.ClassStmt)
	// Calling parent:: in anything other than a class is not allowed.
	if !ok {
		return nil, errors.New("Unexpected call of parent:: outside of a class")
	}

	if classStmt.Extends == nil {
		return nil, errors.New("Unexpected call of parent:: in a class without a parent")
	}

	fqn := p.FQN(root, classStmt.Extends.ClassName)
	node := p.FindNodeInTrie(fqn, ir.KindClassStmt)
	if node == nil {
		return nil, fmt.Errorf("Parent class %s is not indexed", fqn.String())
	}

	return node, nil
}

func (p *Project) findClassLikeSymbol(clsLike *phpdoxer.TypeClassLike) *traversers.TrieNode {
	results := p.symbolTrie.SearchExact(clsLike.Identifier())
	for _, result := range results {
		if result.Namespace == clsLike.Namespace() {
			return result
		}
	}

	return nil
}
