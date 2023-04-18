package server

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
	"github.com/laytan/phpls/internal/config"
	pcontext "github.com/laytan/phpls/internal/context"
	"github.com/laytan/phpls/internal/fqner"
	"github.com/laytan/phpls/internal/project/definition"
	"github.com/laytan/phpls/internal/wrkspc"

	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/laytan/phpls/pkg/lsperrors"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/position"
	"github.com/laytan/phpls/pkg/traversers"
)

func (s *Server) References(
	ctx context.Context,
	params *protocol.ReferenceParams,
) ([]protocol.Location, error) {
	if err := s.isMethodAllowed("Initialize"); err != nil {
		return nil, err
	}

	// TODO: support params.IncludeDeclaration.

	start := time.Now()
	defer func() {
		log.Printf("Retrieving references took %s", time.Since(start))
	}()

	target := position.FromTextDocumentPositionParams(&params.Position, &params.TextDocument)
	references, err := s.references(ctx, target)
	if err != nil {
		log.Printf("[ERROR]: finding references of %v: %v", target, err)
		return nil, lsperrors.ErrRequestFailed(err.Error())
	}

	return references, nil
}

func (s *Server) references(
	ctx context.Context,
	target *position.Position,
) ([]protocol.Location, error) {
	definitions, err := s.project.Definition(target)
	if err != nil {
		return nil, fmt.Errorf("finding definition of symbol to get references of: %w", err)
	}

	if len(definitions) > 1 {
		return nil, fmt.Errorf("multiple definitions found in references call, not supported")
	}

	d := definitions[0]
	content := wrkspc.Current.ContentF(d.Path)
	dpos := position.FromIRPosition(d.Path, content, d.Position.StartPos)

	dctx, err := pcontext.New(dpos)
	if err != nil {
		return nil, fmt.Errorf("Unable to recreate definition context: %w", err)
	}

	// Advance to the node that matches the definition.
	for ok := true; ok; ok = dctx.Advance() {
		if nodeident.Get(dctx.Current()) == d.Identifier {
			break
		}
	}

	switch td := dctx.Current().(type) {
	case *ast.StmtClass, *ast.StmtInterface, *ast.StmtTrait:
		return s.classReferences(ctx, dctx, d), nil
	case *ast.StmtFunction:
		// Cases: global scope (from index)
		// block scope: check block for references

		// tctx, err := pcontext.New(target)
		// if err != nil {
		// 	return nil, fmt.Errorf("creating target context: %w", err)
		// }

		// currScope := tctx.Scope()
		// If the definition is inside the current scope.
		// References will also only be in the current scope.
		if dctx.Scope().GetType() != ast.TypeRoot {
			if position.PosInAst(dctx.Scope().GetPosition(), target) {
				v := traversers.NewFunctionCall(d.Identifier, true)
				tv := traverser.NewTraverser(v)
				dctx.Scope().Accept(tv)

				res := make([]protocol.Location, 0, len(v.Result)+1)
				res = append(res, position.AstToLspLocation(d.Path, td.Name.GetPosition()))
				for _, call := range v.Result {
					res = append(res, position.AstToLspLocation(d.Path, call.Function.GetPosition()))
				}

				return res, nil
			}
		}

		return nil, lsperrors.ErrRequestFailed("Unimplemented: non-local function references are not implemented")

	// // If have local scope, check it for the function.
	// if scopes.Block.GetType() != ast.TypeRoot {
	// 	ft := traversers.NewFunction(toResolve.Identifier)
	// 	tv := traverser.NewTraverser(ft)
	// 	scopes.Block.Accept(tv)
	//
	// 	if ft.Function != nil {
	// 		return &Resolved{
	// 			Node: ft.Function,
	// 			Path: scopes.Path,
	// 		}, typeOfFunc(ft.Function), phprivacy.PrivacyPublic, true
	// 	}
	// }
	//
	// // Check for functions defined in the used namespaces.
	// if def, ok := fqner.FindFullyQualifiedName(scopes.Root, &ast.Name{
	// 	Position: toResolve.Position,
	// 	Parts:    nameParts(toResolve.Identifier),
	// }); ok {
	// 	n := def.ToIRNode(wrkspc.Current.AstF(def.Path))
	// 	return &Resolved{
	// 		Node: n,
	// 		Path: def.Path,
	// 	}, typeOfFunc(n), phprivacy.PrivacyPublic, true
	// }
	//
	// // Check for global functions.
	// key := fqn.New(fqn.PartSeperator + toResolve.Identifier)
	// def, ok := index.Current.Find(key)
	// if !ok {
	// 	log.Println(fmt.Errorf("[expr.functionResolver.Up]: unable to find %s in index", key))
	// 	return nil, nil, 0, false
	// }
	default:
		msg := fmt.Sprintf("Unsupported node type %T for references", dctx.Current())
		log.Println(msg)
		return nil, lsperrors.ErrRequestFailed(msg)
	}
}

func (s *Server) classReferences(
	ctx context.Context,
	pctx *pcontext.Ctx,
	d *definition.Definition,
) []protocol.Location {
	hasErrors := false

	files := make(chan *wrkspc.ParsedFile)
	references := make(chan protocol.Location)

	tvpool := sync.Pool{
		New: func() any {
			return &classReferenceVisitor{references: references}
		},
	}

	done := &atomic.Uint64{}
	total := &atomic.Uint64{}
	stop, err := s.progress.Track(
		ctx,
		func() float64 { return float64(done.Load()) },
		func() float64 { return float64(total.Load()) },
		"finding references",
		time.Millisecond*50,
	)
	if err != nil {
		log.Printf("[ERROR]: starting references progress: %v", err)
	}
	defer func() {
		if err := stop(nil); err != nil {
			log.Printf("[ERROR]: stopping references progress: %v", err)
		}
	}()

	name := fqner.FullyQualifyName(pctx.Root(), pctx.Current())

	log.Printf("[DEBUG]: finding references of %q", name)

	go func() {
		// If the definition is in stubs or in vendor, we need to check everywhere,
		// But, if the definition is in the project files, it can not be used/referenced in vendor or stubs
		// so, there is no need to walk those directories.
		definitionInVendor := strings.Contains(d.Path, "/vendor/")
		// Note: this does not work in tests.
		definitionInStubs := strings.HasPrefix(d.Path, config.Current.StubsPath)
		walkOpts := wrkspc.WalkOptions{
			DoStubs:  definitionInVendor || definitionInStubs,
			DoVendor: definitionInVendor || definitionInStubs,
		}
		log.Println(walkOpts)

		if err := wrkspc.Current.Walk(files, total, walkOpts); err != nil {
			log.Printf(
				"[WARN]: could not index the file content of root %s: %v",
				wrkspc.Current.Root(),
				err,
			)
			hasErrors = true
		}
	}()

	go func() {
		defer close(references)
		wg := sync.WaitGroup{}
		defer wg.Wait()

		// Definition is also a reference.
		var dname ast.Vertex
		switch tn := pctx.Current().(type) {
		case *ast.StmtClass:
			dname = tn.Name
		case *ast.StmtInterface:
			dname = tn.Name
		case *ast.StmtTrait:
			dname = tn.Name
		default:
			panic("unreachable")
		}
		references <- position.AstToLspLocation(d.Path, dname.GetPosition())

		for file := range files {
			file := file
			wg.Add(1)
			go func() {
				defer done.Add(1)
				defer wg.Done()

				// PERF: We don't parse when the class name is not in the file.
				if !strings.Contains(file.Content, name.Name()) {
					return
				}

				root, err := wrkspc.Current.Parse(file.Path, []byte(file.Content))
				if err != nil {
					log.Printf("ERROR: parsing %q: %v", file.Path, err)
					hasErrors = true
					return
				}

				v := tvpool.Get().(*classReferenceVisitor)
				defer tvpool.Put(v)
				v.Reset(file.Path, file.Content, name)
				tv := traverser.NewTraverser(v)
				root.Accept(tv)
			}()
		}
	}()

	accReferences := []protocol.Location{}
	for reference := range references {
		accReferences = append(accReferences, reference)
	}

	if hasErrors {
		log.Println(
			"Parsing the project for references resulted in errors, check the logs for more details",
		)
	}

	return accReferences
}

type classReferenceVisitor struct {
	visitor.Null
	references chan protocol.Location
	names      []ast.Vertex
	fqnv       *fqn.Traverser
	fqn        *fqn.FQN
	path       string
	content    string
}

func (c *classReferenceVisitor) Reset(path string, content string, name *fqn.FQN) {
	if len(c.names) > 0 {
		c.names = c.names[:0]
	}

	c.fqn = name
	c.path = path
	c.content = content
	c.fqnv = fqn.NewTraverser()
}

func (c *classReferenceVisitor) EnterNode(node ast.Vertex) bool {
	c.fqnv.EnterNode(node)
	return true
}

func (c *classReferenceVisitor) NameName(node *ast.Name) {
	c.names = append(c.names, node)
}

func (c *classReferenceVisitor) NameFullyQualified(node *ast.NameFullyQualified) {
	c.names = append(c.names, node)
}

func (c *classReferenceVisitor) NameRelative(node *ast.NameRelative) {
	c.names = append(c.names, node)
}

func (c *classReferenceVisitor) LeaveNode(node ast.Vertex) {
	if node.GetType() != ast.TypeRoot {
		return
	}

	for _, name := range c.names {
		if name == nil {
			continue
		}

		if c.fqnv.ResultFor(name).String() == c.fqn.String() {
			var lastPart ast.Vertex
			switch tn := name.(type) {
			case *ast.Name:
				lastPart = tn.Parts[len(tn.Parts)-1]
			case *ast.NameFullyQualified:
				lastPart = tn.Parts[len(tn.Parts)-1]
			case *ast.NameRelative:
				lastPart = tn.Parts[len(tn.Parts)-1]
			}
			c.references <- position.AstToLspLocation(c.path, lastPart.GetPosition())
		}
	}
}
