package typer

import (
	"fmt"
	"log"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/queue"
	"github.com/laytan/elephp/pkg/resolvequeue"
	"github.com/laytan/elephp/pkg/traversers"
)

// Returns the return type of the method or function, prioritizing phpdoc
// @return over the return type hint.
func (t *typer) Returns(
	root *ir.Root,
	funcOrMeth ir.Node,
	rootRetriever func(n *resolvequeue.Node) (*ir.Root, error),
) phpdoxer.Type {
	kind := ir.GetNodeKind(funcOrMeth)
	if kind != ir.KindClassMethodStmt && kind != ir.KindFunctionStmt {
		panic(fmt.Errorf("Type: %T: %w", funcOrMeth, ErrUnexpectedNodeType))
	}

	if retDoc := findReturnComment(root, funcOrMeth, rootRetriever); retDoc != nil {
		return resolveFQN(root, funcOrMeth, retDoc)
	}

	if retHint := parseTypeHint(funcOrMeth); retHint != nil {
		return resolveFQN(root, funcOrMeth, retHint)
	}

	return nil
}

func findReturnComment(
	root *ir.Root,
	node ir.Node,
	rootRetriever func(n *resolvequeue.Node) (*ir.Root, error),
) phpdoxer.Type {
	cmntQueue := queue.New[phpdoxer.Node]()
	addFuncComments(cmntQueue, node)

	var hasProcessedInheritDoc bool

	for currCmnt := cmntQueue.Dequeue(); currCmnt != nil; currCmnt = cmntQueue.Dequeue() {
		switch currCmnt.Kind() {
		case phpdoxer.KindReturn:
			return currCmnt.(*phpdoxer.NodeReturn).Type

		case phpdoxer.KindInheritDoc:
			// We only need to process one inheritdoc in the chain.
			if hasProcessedInheritDoc {
				continue
			}

			// Regular function don't support inheritdoc.
			method, ok := node.(*ir.ClassMethodStmt)
			if !ok {
				log.Println("regular function with inheritdoc PHPDoc is not allowed")
				continue
			}

			isStatic, isFinal := parseMethodModifiers(method.Modifiers)
			if isFinal {
				log.Println("final method with inheritdoc PHPDoc is not allowed")
				continue
			}

			resolveQueue := resolvequeue.New(rootRetriever, getMethodClass(root, node))
			// Dequeue the current class.
			resolveQueue.Queue.Dequeue()

			for currCls := resolveQueue.Queue.Dequeue(); currCls != nil; currCls = resolveQueue.Queue.Dequeue() {
				result := findMethod(rootRetriever, currCls, isStatic, method.MethodName.Value)
				if result != nil {
					addFuncComments(cmntQueue, result)
				}
			}

			for _, currCls := range resolveQueue.Implements {
				result := findMethod(rootRetriever, currCls, isStatic, method.MethodName.Value)
				if result != nil {
					addFuncComments(cmntQueue, result)
				}
			}

		default:
			continue
		}
	}

	return nil
}

func findMethod(
	rootRetriever func(n *resolvequeue.Node) (*ir.Root, error),
	currCls *resolvequeue.Node,
	isStatic bool,
	name string,
) *ir.ClassMethodStmt {
	root, err := rootRetriever(currCls)
	if err != nil {
		log.Println(err)
		return nil
	}

	mt := newMethodTraverser(currCls.FQN.Name(), name, isStatic)
	root.Walk(mt)

	return mt.Method
}

func newMethodTraverser(className, methodName string, isStatic bool) *traversers.Method {
	if isStatic {
		return traversers.NewMethodStatic(methodName, className, phprivacy.PrivacyPrivate)
	}

	return traversers.NewMethod(methodName, className, phprivacy.PrivacyPrivate)
}

func parseMethodModifiers(identifiers []*ir.Identifier) (isStatic, isFinal bool) {
	for _, ident := range identifiers {
		if ident.Value == "static" {
			isStatic = true
		} else if ident.Value == "final" {
			isFinal = true
		}
	}

	return isStatic, isFinal
}

func getMethodClass(root *ir.Root, node ir.Node) *resolvequeue.Node {
	t := fqn.NewFQNTraverserHandlingKeywords(node)
	root.Walk(t)
	fqn := t.ResultFor(&ir.Name{Value: "self"})

	return &resolvequeue.Node{
		FQN: fqn,
		// TODO: this could be any class-like.
		Kind: ir.KindClassStmt,
	}
}

func addFuncComments(q *queue.Queue[phpdoxer.Node], funcOrMeth ir.Node) {
	comments := NodeComments(funcOrMeth)

	if method, ok := funcOrMeth.(*ir.ClassMethodStmt); ok {
		comments = append(comments, method.Doc.Raw)
	}

	for _, comment := range comments {
		nodes, err := phpdoxer.ParseDoc(comment)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, node := range nodes {
			q.Enqueue(node)
		}
	}
}
