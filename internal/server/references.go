package server

import (
	"context"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
	pcontext "github.com/laytan/phpls/internal/context"
	"github.com/laytan/phpls/internal/fqner"
	"github.com/laytan/phpls/internal/wrkspc"

	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/phpls/pkg/fqn"
	"github.com/laytan/phpls/pkg/lsperrors"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/parsing"
	"github.com/laytan/phpls/pkg/phpversion"
	"github.com/laytan/phpls/pkg/position"
)

// TODO: add config to skip vendor directory.
func (s *Server) References(
	ctx context.Context,
	params *protocol.ReferenceParams,
) ([]protocol.Location, error) {
	if err := s.isMethodAllowed("Initialize"); err != nil {
		return nil, err
	}

	start := time.Now()
	defer func() {
		log.Printf("Retrieving references took %s", time.Since(start))
	}()

	target := position.FromTextDocumentPositionParams(&params.Position, &params.TextDocument)
	definitions, err := s.project.Definition(target)
	if err != nil {
		log.Println(err)
		return nil, lsperrors.ErrRequestFailed(err.Error())
	}

	if len(definitions) > 1 {
		msg := "Multiple definitions found in references call, not supported"
		log.Println(msg)
		return nil, lsperrors.ErrRequestFailed(msg)
	}

	d := definitions[0]
	content := wrkspc.Current.FContentOf(d.Path)
	dpos := position.FromIRPosition(d.Path, content, d.Position.StartPos)

	pctx, err := pcontext.New(dpos)
	if err != nil {
		log.Printf("Unable to recreate definition context: %v", err)
		return nil, lsperrors.ErrRequestFailed(err.Error())
	}
	// Advance to the node that matches the definition.
	for ok := true; ok; ok = pctx.Advance() {
		if nodeident.Get(pctx.Current()) == d.Identifier {
			break
		}
	}

	switch pctx.Current().(type) {
	case *ast.StmtClass:
		hasErrors := false

		files := make(chan *wrkspc.ParsedFile)
		references := make(chan protocol.Location)

		tvpool := sync.Pool{
			New: func() any {
				return &classReferenceVisitor{references: references}
			},
		}

		// TODO: parser should be like configured.
		parser := parsing.New(phpversion.Latest())
		name := fqner.FullyQualifyName(pctx.Root(), pctx.Current())

		log.Printf("[DEBUG]: finding references of %q", name)

		go func() {
			total := &atomic.Uint64{}
			totalDone := make(chan bool, 1)
			if err := wrkspc.Current.Index(files, total, totalDone); err != nil {
				log.Println(
					fmt.Errorf(
						"Could not index the file content of root %s: %w",
						wrkspc.Current.Root(),
						err,
					),
				)
				hasErrors = true
			}
		}()

		go func() {
			defer close(references)
			wg := sync.WaitGroup{}
			defer wg.Wait()
			for file := range files {
				file := file
				wg.Add(1)
				go func() {
					defer func() {
						wg.Done()
						if r := recover(); r != nil {
							log.Printf("ERROR: could not parse %q into an AST: %v", file.Path, r)
						}
					}()

					root, err := parser.Parse([]byte(file.Content))
					if err != nil {
						log.Printf("ERROR: parsing %q: %v", file.Path, err)
						hasErrors = true
						return
					}

					// TODO: sync.pool
					v := tvpool.Get().(*classReferenceVisitor)
					defer tvpool.Put(v)
					v.Reset(file.Path, file.Content, name)
					tv := traverser.NewTraverser(v)
					root.Accept(tv)
				}()
			}
		}()

		var accReferences []protocol.Location
		for reference := range references {
			log.Printf("[DEBUG]: found reference: %v", reference)
			accReferences = append(accReferences, reference)
		}

		if hasErrors {
			log.Println(
				"Parsing the project for references resulted in errors, check the logs for more details",
			)
		}

		log.Printf("[DEBUG]: References: %v", accReferences)

		return accReferences, nil
	default:
		// TODO: others.
		msg := fmt.Sprintf("Unsupported node type %T for references", pctx.Current())
		log.Println(msg)
		return nil, lsperrors.ErrRequestFailed(msg)
	}
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

func (c *classReferenceVisitor) StmtClass(node *ast.StmtClass) {
	c.names = append(c.names, node.Name)
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
		if c.fqnv.ResultFor(name).String() == c.fqn.String() {
			c.references <- position.AstToLspLocation(c.path, name.GetPosition())
		}
	}
}
