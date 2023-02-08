package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/laytan/elephp/pkg/position"
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
			Label: res.Symbol.Identifier(),
			Data:  string(pathData),
		}

		if res.FQN.Namespace() == "" {
			items = append(items, item)
			continue
		}

		switch res.Symbol.NodeKind() {
		case ir.KindFunctionStmt:
			item.Kind = protocol.FunctionCompletion
		case ir.KindClassStmt:
			item.Kind = protocol.ClassCompletion
		case ir.KindTraitStmt:
			// NOTE: there is no trait kind, module seems appropriate.
			item.Kind = protocol.ModuleCompletion
		case ir.KindInterfaceStmt:
			item.Kind = protocol.InterfaceCompletion
		default:
		}

		item.Detail = fmt.Sprintf(`%s\%s`, res.FQN.Namespace(), res.Symbol.Identifier())

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
func (s *Server) Resolve(
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
	item.Documentation = s.documentation(item, pos)

	return item, nil
}

func (s *Server) additionalTextEdits(
	item *protocol.CompletionItem,
	pos *position.Position,
) []protocol.TextEdit {
	// get FQN for item, if it matches the namespace in item.Detail, it is already imported.
	if !s.project.NeedsUseStmtFor(pos, `\`+item.Detail) {
		return nil
	}

	// TODO: add use function or use const if it is a function or const.
	// TODO: add alias when there is already a class used with the same name.
	useStmt := fmt.Sprintf("use %s;", item.Detail)
	nsPos := s.project.Namespace(pos)
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

// TODO: once hover is added, add the hover output here.
func (s *Server) documentation(item *protocol.CompletionItem, pos *position.Position) string {
	return ""
}

// This looks like a duplicate of Resolve, what's the difference?
func (s *Server) ResolveCompletionItem(
	context.Context,
	*protocol.CompletionItem,
) (*protocol.CompletionItem, error) {
	return nil, errorUnimplemented
}
