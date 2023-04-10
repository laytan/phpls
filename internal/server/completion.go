package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/php-parser/pkg/ast"
)

func (s *Server) Completion(
	ctx context.Context,
	params *protocol.CompletionParams,
) (*protocol.CompletionList, error) {
	start := time.Now()
	defer func() { log.Printf("Retrieving completion took %s\n", time.Since(start)) }()

	pos := position.FromTextDocumentPositionParams(&params.Position, &params.TextDocument)

	// Turn the position into JSON so we can access it in the Resolve request.
	pathData, err := json.Marshal(pos)
	if err != nil {
		log.Println(err)
		return nil, lsperrors.ErrRequestFailed("Could not serialize document position to JSON")
	}

	results := s.project.Complete(pos)

	items := make([]protocol.CompletionItem, 0, len(results))
	for _, res := range results {
		item := protocol.CompletionItem{
			Label: res.Identifier,
			Data:  string(pathData),
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
			// NOTE: there is no trait kind, module seems appropriate.
			item.Kind = protocol.ModuleCompletion
		case ast.TypeStmtInterface:
			item.Kind = protocol.InterfaceCompletion
		default:
		}

		item.Detail = fmt.Sprintf(`%s\%s`, res.FQN.Namespace(), res.Identifier)

		// NOTE: adding an additional text edit, so the client shows it in the UI.
		// NOTE: the actual text edit is added in the Resolve method.
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
		"Completion item data property could not be converted back to a Position struct, trying to convert: %#v",
		item.Data,
	)

	rawPos, ok := item.Data.(string)
	if !ok {
		log.Println(errMsg)
		return nil, lsperrors.ErrRequestFailed(errMsg)
	}

	pos := &position.Position{}
	err := json.Unmarshal([]byte(rawPos), pos)
	if err != nil {
		log.Println(errMsg)
		log.Println(err)
		return nil, lsperrors.ErrRequestFailed(errMsg)
	}

	item.AdditionalTextEdits = s.additionalTextEdits(item, pos)
	item.Documentation = &protocol.Or_CompletionItem_documentation{
		Value: s.documentation(item, pos),
	}

	return item, nil
}

func (s *Server) additionalTextEdits(
	item *protocol.CompletionItem,
	pos *position.Position,
) []protocol.TextEdit {
	return s.useInserter.Insert(fqn.New(`\`+item.Detail), pos)
}

func (s *Server) documentation(item *protocol.CompletionItem, pos *position.Position) string {
	return ""
}

type UseInserter struct {
	Project *project.Project
}

func (u *UseInserter) Insert(
	qualifiedName *fqn.FQN,
	currPos *position.Position,
) []protocol.TextEdit {
	if !u.Project.NeedsUseStmtFor(currPos, qualifiedName.String()) {
		return nil
	}

	useStmt := fmt.Sprintf("use %s;", qualifiedName.String()[1:])
	nsPos := u.Project.Namespace(currPos)
	if nsPos == nil {
		return []protocol.TextEdit{{
			Range: protocol.Range{
				Start: protocol.Position{Line: 1},
				End:   protocol.Position{Line: 1},
			},
			NewText: useStmt + "\n",
		}}
	}

	retPos := nsPos.ToLSPLocation()

	return []protocol.TextEdit{{
		Range: protocol.Range{
			Start: protocol.Position{Line: retPos.Range.Start.Line + 2},
			End:   protocol.Position{Line: retPos.Range.Start.Line + 2},
		},
		NewText: useStmt + "\n",
	}}
}
