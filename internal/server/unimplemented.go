package server

import (
	"context"

	"github.com/laytan/go-lsp-protocol/pkg/jsonrpc2"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
)

var errorUnimplemented = jsonrpc2.ErrMethodNotFound

func (s *Server) DidChangeWorkspaceFolders(
	context.Context,
	*protocol.DidChangeWorkspaceFoldersParams,
) error {
	return errorUnimplemented
}

func (s *Server) DidChangeNotebookDocument(
	context.Context,
	*protocol.DidChangeNotebookDocumentParams,
) error {
	return errorUnimplemented
}

func (s *Server) DidCloseNotebookDocument(
	context.Context,
	*protocol.DidCloseNotebookDocumentParams,
) error {
	return errorUnimplemented
}

func (s *Server) DidOpenNotebookDocument(
	context.Context,
	*protocol.DidOpenNotebookDocumentParams,
) error {
	return errorUnimplemented
}

func (s *Server) DidSaveNotebookDocument(
	context.Context,
	*protocol.DidSaveNotebookDocumentParams,
) error {
	return errorUnimplemented
}

func (s *Server) WorkDoneProgressCancel(
	context.Context,
	*protocol.WorkDoneProgressCancelParams,
) error {
	return errorUnimplemented
}

func (s *Server) InlayHint(
	context.Context,
	*protocol.InlayHintParams,
) ([]protocol.InlayHint, error) {
	return nil, errorUnimplemented
}

func (s *Server) InlineValue(
	context.Context,
	*protocol.InlineValueParams,
) ([]protocol.Or_InlineValue, error) {
	return nil, errorUnimplemented
}

func (s *Server) Progress(context.Context, *protocol.ProgressParams) error {
	return errorUnimplemented
}

func (s *Server) Resolve(context.Context, *protocol.InlayHint) (*protocol.InlayHint, error) {
	return nil, errorUnimplemented
}

func (s *Server) DidCreateFiles(context.Context, *protocol.CreateFilesParams) error {
	return errorUnimplemented
}

func (s *Server) DidRenameFiles(context.Context, *protocol.RenameFilesParams) error {
	return errorUnimplemented
}

func (s *Server) DidDeleteFiles(context.Context, *protocol.DeleteFilesParams) error {
	return errorUnimplemented
}

func (s *Server) DidChangeConfiguration(
	context.Context,
	*protocol.DidChangeConfigurationParams,
) error {
	return errorUnimplemented
}

func (s *Server) DidSave(context.Context, *protocol.DidSaveTextDocumentParams) error {
	return errorUnimplemented
}

func (s *Server) WillSave(context.Context, *protocol.WillSaveTextDocumentParams) error {
	return errorUnimplemented
}

func (s *Server) DidChangeWatchedFiles(
	context.Context,
	*protocol.DidChangeWatchedFilesParams,
) error {
	return errorUnimplemented
}

func (s *Server) SetTrace(context.Context, *protocol.SetTraceParams) error {
	return errorUnimplemented
}

func (s *Server) LogTrace(context.Context, *protocol.LogTraceParams) error {
	return errorUnimplemented
}

func (s *Server) Implementation(
	context.Context,
	*protocol.ImplementationParams,
) ([]protocol.Location, error) {
	return nil, errorUnimplemented
}

func (s *Server) TypeDefinition(
	context.Context,
	*protocol.TypeDefinitionParams,
) ([]protocol.Location, error) {
	return nil, errorUnimplemented
}

func (s *Server) DocumentColor(
	context.Context,
	*protocol.DocumentColorParams,
) ([]protocol.ColorInformation, error) {
	return nil, errorUnimplemented
}

func (s *Server) ColorPresentation(
	context.Context,
	*protocol.ColorPresentationParams,
) ([]protocol.ColorPresentation, error) {
	return nil, errorUnimplemented
}

func (s *Server) FoldingRange(
	context.Context,
	*protocol.FoldingRangeParams,
) ([]protocol.FoldingRange, error) {
	return nil, errorUnimplemented
}

func (s *Server) Declaration(
	context.Context,
	*protocol.DeclarationParams,
) (*protocol.Or_textDocument_declaration, error) {
	return nil, errorUnimplemented
}

func (s *Server) SelectionRange(
	context.Context,
	*protocol.SelectionRangeParams,
) ([]protocol.SelectionRange, error) {
	return nil, errorUnimplemented
}

func (s *Server) PrepareCallHierarchy(
	context.Context,
	*protocol.CallHierarchyPrepareParams,
) ([]protocol.CallHierarchyItem, error) {
	return nil, errorUnimplemented
}

func (s *Server) IncomingCalls(
	context.Context,
	*protocol.CallHierarchyIncomingCallsParams,
) ([]protocol.CallHierarchyIncomingCall, error) {
	return nil, errorUnimplemented
}

func (s *Server) OutgoingCalls(
	context.Context,
	*protocol.CallHierarchyOutgoingCallsParams,
) ([]protocol.CallHierarchyOutgoingCall, error) {
	return nil, errorUnimplemented
}

func (s *Server) SemanticTokensFull(
	context.Context,
	*protocol.SemanticTokensParams,
) (*protocol.SemanticTokens, error) {
	return nil, errorUnimplemented
}

func (s *Server) SemanticTokensFullDelta(
	context.Context,
	*protocol.SemanticTokensDeltaParams,
) (any, error) {
	return nil, errorUnimplemented
}

func (s *Server) SemanticTokensRange(
	context.Context,
	*protocol.SemanticTokensRangeParams,
) (*protocol.SemanticTokens, error) {
	return nil, errorUnimplemented
}

func (s *Server) SemanticTokensRefresh(context.Context) error {
	return errorUnimplemented
}

func (s *Server) LinkedEditingRange(
	context.Context,
	*protocol.LinkedEditingRangeParams,
) (*protocol.LinkedEditingRanges, error) {
	return nil, errorUnimplemented
}

func (s *Server) WillCreateFiles(
	context.Context,
	*protocol.CreateFilesParams,
) (*protocol.WorkspaceEdit, error) {
	return nil, errorUnimplemented
}

func (s *Server) WillRenameFiles(
	context.Context,
	*protocol.RenameFilesParams,
) (*protocol.WorkspaceEdit, error) {
	return nil, errorUnimplemented
}

func (s *Server) WillDeleteFiles(
	context.Context,
	*protocol.DeleteFilesParams,
) (*protocol.WorkspaceEdit, error) {
	return nil, errorUnimplemented
}

func (s *Server) Moniker(context.Context, *protocol.MonikerParams) ([]protocol.Moniker, error) {
	return nil, errorUnimplemented
}

func (s *Server) PrepareTypeHierarchy(
	context.Context,
	*protocol.TypeHierarchyPrepareParams,
) ([]protocol.TypeHierarchyItem, error) {
	return nil, errorUnimplemented
}

func (s *Server) Supertypes(
	context.Context,
	*protocol.TypeHierarchySupertypesParams,
) ([]protocol.TypeHierarchyItem, error) {
	return nil, errorUnimplemented
}

func (s *Server) Subtypes(
	context.Context,
	*protocol.TypeHierarchySubtypesParams,
) ([]protocol.TypeHierarchyItem, error) {
	return nil, errorUnimplemented
}

func (s *Server) InlayHintRefresh(context.Context) error {
	return errorUnimplemented
}

func (s *Server) WillSaveWaitUntil(
	context.Context,
	*protocol.WillSaveTextDocumentParams,
) ([]protocol.TextEdit, error) {
	return nil, errorUnimplemented
}

func (s *Server) SignatureHelp(
	context.Context,
	*protocol.SignatureHelpParams,
) (*protocol.SignatureHelp, error) {
	return nil, errorUnimplemented
}

func (s *Server) References(
	context.Context,
	*protocol.ReferenceParams,
) ([]protocol.Location, error) {
	return nil, errorUnimplemented
}

func (s *Server) DocumentHighlight(
	context.Context,
	*protocol.DocumentHighlightParams,
) ([]protocol.DocumentHighlight, error) {
	return nil, errorUnimplemented
}

func (s *Server) DocumentSymbol(
	context.Context,
	*protocol.DocumentSymbolParams,
) ([]any, error) {
	return nil, errorUnimplemented
}

func (s *Server) CodeAction(
	context.Context,
	*protocol.CodeActionParams,
) ([]protocol.CodeAction, error) {
	return nil, errorUnimplemented
}

func (s *Server) ResolveCodeAction(
	context.Context,
	*protocol.CodeAction,
) (*protocol.CodeAction, error) {
	return nil, errorUnimplemented
}

func (s *Server) Symbol(
	context.Context,
	*protocol.WorkspaceSymbolParams,
) ([]protocol.SymbolInformation, error) {
	return nil, errorUnimplemented
}

func (s *Server) ResolveWorkspaceSymbol(
	context.Context,
	*protocol.WorkspaceSymbol,
) (*protocol.WorkspaceSymbol, error) {
	return nil, errorUnimplemented
}

func (s *Server) CodeLens(context.Context, *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	return nil, errorUnimplemented
}

func (s *Server) ResolveCodeLens(context.Context, *protocol.CodeLens) (*protocol.CodeLens, error) {
	return nil, errorUnimplemented
}

func (s *Server) CodeLensRefresh(context.Context) error {
	return errorUnimplemented
}

func (s *Server) DocumentLink(
	context.Context,
	*protocol.DocumentLinkParams,
) ([]protocol.DocumentLink, error) {
	return nil, errorUnimplemented
}

func (s *Server) ResolveDocumentLink(
	context.Context,
	*protocol.DocumentLink,
) (*protocol.DocumentLink, error) {
	return nil, errorUnimplemented
}

func (s *Server) RangeFormatting(
	context.Context,
	*protocol.DocumentRangeFormattingParams,
) ([]protocol.TextEdit, error) {
	return nil, errorUnimplemented
}

func (s *Server) OnTypeFormatting(
	context.Context,
	*protocol.DocumentOnTypeFormattingParams,
) ([]protocol.TextEdit, error) {
	return nil, errorUnimplemented
}

func (s *Server) Rename(context.Context, *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	return nil, errorUnimplemented
}

func (s *Server) PrepareRename(
	context.Context,
	*protocol.PrepareRenameParams,
) (*protocol.Msg_PrepareRename2Gn, error) {
	return nil, errorUnimplemented
}

func (s *Server) ExecuteCommand(
	context.Context,
	*protocol.ExecuteCommandParams,
) (any, error) {
	return nil, errorUnimplemented
}

func (s *Server) Diagnostic(context.Context, *string) (*string, error) {
	return nil, errorUnimplemented
}

func (s *Server) DiagnosticWorkspace(
	context.Context,
	*protocol.WorkspaceDiagnosticParams,
) (*protocol.WorkspaceDiagnosticReport, error) {
	return nil, errorUnimplemented
}

func (s *Server) DiagnosticRefresh(context.Context) error {
	return errorUnimplemented
}

func (s *Server) NonstandardRequest(
	ctx context.Context,
	method string,
	params any,
) (any, error) {
	return nil, errorUnimplemented
}
