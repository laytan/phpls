package server

import (
	"context"

	"github.com/jdbaldry/go-language-server-protocol/jsonrpc2"
	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
)

var errorUnimplemented = jsonrpc2.ErrMethodNotFound

func (s *server) DidChangeWorkspaceFolders(
	context.Context,
	*protocol.DidChangeWorkspaceFoldersParams,
) error {
	return errorUnimplemented
}

func (s *server) WorkDoneProgressCancel(
	context.Context,
	*protocol.WorkDoneProgressCancelParams,
) error {
	return errorUnimplemented
}

func (s *server) DidCreateFiles(context.Context, *protocol.CreateFilesParams) error {
	return errorUnimplemented
}

func (s *server) DidRenameFiles(context.Context, *protocol.RenameFilesParams) error {
	return errorUnimplemented
}

func (s *server) DidDeleteFiles(context.Context, *protocol.DeleteFilesParams) error {
	return errorUnimplemented
}

func (s *server) DidChangeConfiguration(
	context.Context,
	*protocol.DidChangeConfigurationParams,
) error {
	return errorUnimplemented
}

func (s *server) DidSave(context.Context, *protocol.DidSaveTextDocumentParams) error {
	return errorUnimplemented
}

func (s *server) WillSave(context.Context, *protocol.WillSaveTextDocumentParams) error {
	return errorUnimplemented
}

func (s *server) DidChangeWatchedFiles(
	context.Context,
	*protocol.DidChangeWatchedFilesParams,
) error {
	return errorUnimplemented
}

func (s *server) SetTrace(context.Context, *protocol.SetTraceParams) error {
	return errorUnimplemented
}

func (s *server) LogTrace(context.Context, *protocol.LogTraceParams) error {
	return errorUnimplemented
}

func (s *server) Implementation(
	context.Context,
	*protocol.ImplementationParams,
) (protocol.Definition, error) {
	return nil, errorUnimplemented
}

func (s *server) TypeDefinition(
	context.Context,
	*protocol.TypeDefinitionParams,
) (protocol.Definition, error) {
	return nil, errorUnimplemented
}

func (s *server) DocumentColor(
	context.Context,
	*protocol.DocumentColorParams,
) ([]protocol.ColorInformation, error) {
	return nil, errorUnimplemented
}

func (s *server) ColorPresentation(
	context.Context,
	*protocol.ColorPresentationParams,
) ([]protocol.ColorPresentation, error) {
	return nil, errorUnimplemented
}

func (s *server) FoldingRange(
	context.Context,
	*protocol.FoldingRangeParams,
) ([]protocol.FoldingRange, error) {
	return nil, errorUnimplemented
}

func (s *server) Declaration(
	context.Context,
	*protocol.DeclarationParams,
) (protocol.Declaration, error) {
	return nil, errorUnimplemented
}

func (s *server) SelectionRange(
	context.Context,
	*protocol.SelectionRangeParams,
) ([]protocol.SelectionRange, error) {
	return nil, errorUnimplemented
}

func (s *server) PrepareCallHierarchy(
	context.Context,
	*protocol.CallHierarchyPrepareParams,
) ([]protocol.CallHierarchyItem, error) {
	return nil, errorUnimplemented
}

func (s *server) IncomingCalls(
	context.Context,
	*protocol.CallHierarchyIncomingCallsParams,
) ([]protocol.CallHierarchyIncomingCall, error) {
	return nil, errorUnimplemented
}

func (s *server) OutgoingCalls(
	context.Context,
	*protocol.CallHierarchyOutgoingCallsParams,
) ([]protocol.CallHierarchyOutgoingCall, error) {
	return nil, errorUnimplemented
}

func (s *server) SemanticTokensFull(
	context.Context,
	*protocol.SemanticTokensParams,
) (*protocol.SemanticTokens, error) {
	return nil, errorUnimplemented
}

func (s *server) SemanticTokensFullDelta(
	context.Context,
	*protocol.SemanticTokensDeltaParams,
) (interface{}, error) {
	return nil, errorUnimplemented
}

func (s *server) SemanticTokensRange(
	context.Context,
	*protocol.SemanticTokensRangeParams,
) (*protocol.SemanticTokens, error) {
	return nil, errorUnimplemented
}

func (s *server) SemanticTokensRefresh(context.Context) error {
	return errorUnimplemented
}

func (s *server) LinkedEditingRange(
	context.Context,
	*protocol.LinkedEditingRangeParams,
) (*protocol.LinkedEditingRanges, error) {
	return nil, errorUnimplemented
}

func (s *server) WillCreateFiles(
	context.Context,
	*protocol.CreateFilesParams,
) (*protocol.WorkspaceEdit, error) {
	return nil, errorUnimplemented
}

func (s *server) WillRenameFiles(
	context.Context,
	*protocol.RenameFilesParams,
) (*protocol.WorkspaceEdit, error) {
	return nil, errorUnimplemented
}

func (s *server) WillDeleteFiles(
	context.Context,
	*protocol.DeleteFilesParams,
) (*protocol.WorkspaceEdit, error) {
	return nil, errorUnimplemented
}

func (s *server) Moniker(context.Context, *protocol.MonikerParams) ([]protocol.Moniker, error) {
	return nil, errorUnimplemented
}

func (s *server) PrepareTypeHierarchy(
	context.Context,
	*protocol.TypeHierarchyPrepareParams,
) ([]protocol.TypeHierarchyItem, error) {
	return nil, errorUnimplemented
}

func (s *server) Supertypes(
	context.Context,
	*protocol.TypeHierarchySupertypesParams,
) ([]protocol.TypeHierarchyItem, error) {
	return nil, errorUnimplemented
}

func (s *server) Subtypes(
	context.Context,
	*protocol.TypeHierarchySubtypesParams,
) ([]protocol.TypeHierarchyItem, error) {
	return nil, errorUnimplemented
}

func (s *server) Resolve(
	context.Context,
	*protocol.CompletionItem,
) (*protocol.CompletionItem, error) {
	return nil, errorUnimplemented
}

func (s *server) InlayHintRefresh(context.Context) error {
	return errorUnimplemented
}

func (s *server) WillSaveWaitUntil(
	context.Context,
	*protocol.WillSaveTextDocumentParams,
) ([]protocol.TextEdit, error) {
	return nil, errorUnimplemented
}

func (s *server) Completion(
	context.Context,
	*protocol.CompletionParams,
) (*protocol.CompletionList, error) {
	return nil, errorUnimplemented
}

func (s *server) ResolveCompletionItem(
	context.Context,
	*protocol.CompletionItem,
) (*protocol.CompletionItem, error) {
	return nil, errorUnimplemented
}

func (s *server) Hover(context.Context, *protocol.HoverParams) (*protocol.Hover, error) {
	return nil, errorUnimplemented
}

func (s *server) SignatureHelp(
	context.Context,
	*protocol.SignatureHelpParams,
) (*protocol.SignatureHelp, error) {
	return nil, errorUnimplemented
}

func (s *server) References(
	context.Context,
	*protocol.ReferenceParams,
) ([]protocol.Location, error) {
	return nil, errorUnimplemented
}

func (s *server) DocumentHighlight(
	context.Context,
	*protocol.DocumentHighlightParams,
) ([]protocol.DocumentHighlight, error) {
	return nil, errorUnimplemented
}

func (s *server) DocumentSymbol(
	context.Context,
	*protocol.DocumentSymbolParams,
) ([]interface{}, error) {
	return nil, errorUnimplemented
}

func (s *server) CodeAction(
	context.Context,
	*protocol.CodeActionParams,
) ([]protocol.CodeAction, error) {
	return nil, errorUnimplemented
}

func (s *server) ResolveCodeAction(
	context.Context,
	*protocol.CodeAction,
) (*protocol.CodeAction, error) {
	return nil, errorUnimplemented
}

func (s *server) Symbol(
	context.Context,
	*protocol.WorkspaceSymbolParams,
) ([]protocol.SymbolInformation, error) {
	return nil, errorUnimplemented
}

func (s *server) CodeLens(context.Context, *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	return nil, errorUnimplemented
}

func (s *server) ResolveCodeLens(context.Context, *protocol.CodeLens) (*protocol.CodeLens, error) {
	return nil, errorUnimplemented
}

func (s *server) CodeLensRefresh(context.Context) error {
	return errorUnimplemented
}

func (s *server) DocumentLink(
	context.Context,
	*protocol.DocumentLinkParams,
) ([]protocol.DocumentLink, error) {
	return nil, errorUnimplemented
}

func (s *server) ResolveDocumentLink(
	context.Context,
	*protocol.DocumentLink,
) (*protocol.DocumentLink, error) {
	return nil, errorUnimplemented
}

func (s *server) Formatting(
	context.Context,
	*protocol.DocumentFormattingParams,
) ([]protocol.TextEdit, error) {
	return nil, errorUnimplemented
}

func (s *server) RangeFormatting(
	context.Context,
	*protocol.DocumentRangeFormattingParams,
) ([]protocol.TextEdit, error) {
	return nil, errorUnimplemented
}

func (s *server) OnTypeFormatting(
	context.Context,
	*protocol.DocumentOnTypeFormattingParams,
) ([]protocol.TextEdit, error) {
	return nil, errorUnimplemented
}

func (s *server) Rename(context.Context, *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	return nil, errorUnimplemented
}

func (s *server) PrepareRename(
	context.Context,
	*protocol.PrepareRenameParams,
) (*protocol.Range, error) {
	return nil, errorUnimplemented
}

func (s *server) ExecuteCommand(
	context.Context,
	*protocol.ExecuteCommandParams,
) (interface{}, error) {
	return nil, errorUnimplemented
}

func (s *server) Diagnostic(context.Context, *string) (*string, error) {
	return nil, errorUnimplemented
}

func (s *server) DiagnosticWorkspace(
	context.Context,
	*protocol.WorkspaceDiagnosticParams,
) (*protocol.WorkspaceDiagnosticReport, error) {
	return nil, errorUnimplemented
}

func (s *server) DiagnosticRefresh(context.Context) error {
	return errorUnimplemented
}

func (s *server) NonstandardRequest(
	ctx context.Context,
	method string,
	params interface{},
) (interface{}, error) {
	return nil, errorUnimplemented
}
