package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/laytan/elephp/internal/fqner"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/laytan/elephp/pkg/nodescopes"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
)

type CompletionItemData struct {
	CurrPos   *position.Position
	TargetPos *position.Position
}

func (s *Server) Completion(
	ctx context.Context,
	params *protocol.CompletionParams,
) (*protocol.CompletionList, error) {
	start := time.Now()
	defer func() { log.Printf("Retrieving completion took %s\n", time.Since(start)) }()

	pos := position.FromTextDocumentPositionParams(&params.Position, &params.TextDocument)
	completionData := func(n *index.INode) string {
		data := CompletionItemData{
			CurrPos: pos,
			TargetPos: &position.Position{
				Path: n.Path,
				Row:  uint(n.Position.StartLine),
				Col:  uint(n.Position.StartCol),
			},
		}
		dataJSON, err := json.Marshal(data)
		if err != nil {
			panic(err) // Should never fail.
		}
		return string(dataJSON)
	}

	results := s.project.Complete(pos)
	items := make([]protocol.CompletionItem, 0, len(results))
	for _, res := range results {
		item := protocol.CompletionItem{
			Label: res.Identifier,
			Data:  completionData(res),
		}

		if res.FQN.Namespace() == "" {
			items = append(items, item)
			continue
		}

		switch res.Kind {
		case ast.TypeStmtFunction:
			item.Kind = protocol.FunctionCompletion
		case ast.TypeStmtClass:
			item.Kind = protocol.ClassCompletion
		case ast.TypeStmtTrait:
			// There is no trait kind, module seems appropriate.
			item.Kind = protocol.ModuleCompletion
		case ast.TypeStmtInterface:
			item.Kind = protocol.InterfaceCompletion
		default:
		}

		item.Detail = fmt.Sprintf(`%s\%s`, res.FQN.Namespace(), res.Identifier)

		// Adding an additional text edit, so the client shows it in the UI.
		// The actual text edit is added in the Resolve method.
		item.AdditionalTextEdits = []protocol.TextEdit{{}}

		items = append(items, item)
	}

	log.Printf("Returning %d completion items\n", len(items))

	return &protocol.CompletionList{
		Items:        items,
		IsIncomplete: true,
	}, nil
}

// Completion will return basic information about the result,
// Resolve adds to that with documentation, importing etc.
// This is called when the completion item from Complete is selected/hovered.
func (s *Server) ResolveCompletionItem(
	ctx context.Context,
	item *protocol.CompletionItem,
) (*protocol.CompletionItem, error) {
	start := time.Now()
	defer func() { log.Printf("Resolving completion took %s\n", time.Since(start)) }()

	errMsg := fmt.Sprintf(
		"Completion item data property could not be converted back to a CompletionItemData struct, trying to convert: %#v",
		item.Data,
	)

	rawData, ok := item.Data.(string)
	if !ok {
		log.Println(errMsg)
		return nil, lsperrors.ErrRequestFailed(errMsg)
	}

	var data CompletionItemData
	err := json.Unmarshal([]byte(rawData), &data)
	if err != nil {
		log.Println(errMsg)
		log.Println(err)
		return nil, lsperrors.ErrRequestFailed(errMsg)
	}

	item.AdditionalTextEdits = s.additionalTextEdits(item, data.CurrPos)
	item.Documentation = &protocol.Or_CompletionItem_documentation{
		Value: s.documentation(data.TargetPos),
	}

	return item, nil
}

func (s *Server) additionalTextEdits(
	item *protocol.CompletionItem,
	pos *position.Position,
) []protocol.TextEdit {
	return InsertUseStmt(fqn.New(`\`+item.Detail), pos)
}

// This works but doesn't really look good, but that is probably a problem with the Hover method.
func (s *Server) documentation(pos *position.Position) string {
	return s.project.Hover(pos)
}

func InsertUseStmt(
	qualifiedName *fqn.FQN,
	currPos *position.Position,
) []protocol.TextEdit {
	if !fqner.NeedsUseStmtFor(currPos, qualifiedName) {
		return nil
	}

	root := wrkspc.Current.FIROf(currPos.Path)
	v := &useInserterVisitor{EndLine: int(currPos.Row)}
	tv := traverser.NewTraverser(v)
	root.Accept(tv)

	line := 1
	text := fmt.Sprintf("use %s;\n", qualifiedName.String()[1:])
	if v.ResultLine != 0 {
		line = v.ResultLine
		if v.ShouldPrependLinebreak {
			text = "\n" + text
		}
	}

	return []protocol.TextEdit{{
		Range: protocol.Range{
			Start: protocol.Position{Line: uint32(line)},
			End:   protocol.Position{Line: uint32(line)},
		},
		NewText: text,
	}}
}

type useInserterVisitor struct {
	visitor.Null

	// Line to stop looking after. Should be the line at which insertion is asked.
	EndLine                int
	ResultLine             int
	ShouldPrependLinebreak bool
}

func (u *useInserterVisitor) EnterNode(node ast.Vertex) bool {
	// Stop after given row.
	if node.GetPosition().StartLine >= u.EndLine {
		return false
	}

	// Don't go into scopes, namespace and use is always top level.
	return !nodescopes.IsScope(node.GetType())
}

func (u *useInserterVisitor) Root(node *ast.Root) {
	if len(node.Stmts) > 0 {
		if html, ok := node.Stmts[0].(*ast.StmtInlineHtml); ok {
			u.ResultLine = html.Position.EndLine + 1
			u.ShouldPrependLinebreak = true
		}
	}
}

func (u *useInserterVisitor) StmtNamespace(node *ast.StmtNamespace) {
	// This is a block namespace, like: namespace Test { ...stmts }.
	if node.OpenCurlyBracketTkn != nil {
		u.ResultLine = node.Position.StartLine
		u.ShouldPrependLinebreak = true
	} else {
		u.ResultLine = node.Position.EndLine
		u.ShouldPrependLinebreak = true
	}
}

func (u *useInserterVisitor) StmtUseDeclaration(node *ast.StmtUse) {
	u.ResultLine = node.Position.EndLine
	u.ShouldPrependLinebreak = false
}
