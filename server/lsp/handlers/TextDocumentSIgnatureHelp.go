package handlers

import (
	"strings"

	"github.com/pherrymason/c3-lsp/lsp/symbols"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// textDocument/signatureHelp: {"context":{"isRetrigger":false,"triggerCharacter":"(","triggerKind":2},"position":{"character":20,"line":8},"textDocument":{"uri":"file:///Volumes/Development/raul/projects/game-dev/raul-game-project/murder-c3/src/main.c3"}}
func (h *Handlers) TextDocumentSignatureHelp(context *glsp.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	doc, ok := h.documents.Get(params.TextDocument.URI)
	if !ok {
		return nil, nil
	}

	// Rewind position after previous "("
	posOption := doc.SourceCode.RewindBeforePreviousParenthesis(symbols.NewPositionFromLSPPosition(params.Position))

	if posOption.IsNone() {
		return nil, nil
	}

	foundSymbolOption := h.language.FindSymbolDeclarationInWorkspace(doc, posOption.Get())
	if foundSymbolOption.IsNone() {
		return nil, nil
	}

	foundSymbol := foundSymbolOption.Get()
	function, ok := foundSymbol.(*symbols.Function)
	if !ok {
		return nil, nil
	}

	parameters := []protocol.ParameterInformation{}
	argsToStringify := []string{}
	for _, arg := range function.GetArguments() {
		argsToStringify = append(
			argsToStringify,
			arg.GetType().String()+" "+arg.GetName(),
		)
		parameters = append(
			parameters,
			protocol.ParameterInformation{
				Label: arg.GetType().String() + " " + arg.GetName(),
			},
		)
	}

	signatureHelp := protocol.SignatureHelp{
		Signatures: []protocol.SignatureInformation{
			{
				Label:         function.GetFQN() + "(" + strings.Join(argsToStringify, ", ") + ")",
				Parameters:    parameters,
				Documentation: "Blbalbal bals ldba sdadfa isvaids v",
			},
		},
	}

	return &signatureHelp, nil
}
