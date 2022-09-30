package position

import (
	"bufio"
	"strings"

	"github.com/jdbaldry/go-language-server-protocol/lsp/protocol"
)

// file://
const URIFilePrefixLength = 7

type Position struct {
	Row  uint
	Col  uint
	Path string
}

func FromTextDocumentPositionParams(
	pos *protocol.Position,
	doc *protocol.TextDocumentIdentifier,
) *Position {
	return &Position{
		Row:  uint(pos.Line + 1),
		Col:  uint(pos.Character + 1),
		Path: string(doc.URI)[URIFilePrefixLength:],
	}
}

func (p *Position) ToLSPLocation() protocol.Location {
	return protocol.Location{
		URI: protocol.DocumentURI("file://" + p.Path),
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(p.Row) - 1,
				Character: uint32(p.Col) - 1,
			},
			End: protocol.Position{
				Line:      uint32(p.Row) - 1,
				Character: uint32(p.Col) - 1,
			},
		},
	}
}

func PosToLoc(content string, pos uint) (row uint, col uint) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	linebreaks := 1
	if strings.Contains(content, "\r\n") {
		linebreaks = 2
	}

	var line uint
	var curr uint
	for scanner.Scan() {
		line++

		lineLen := uint(len(scanner.Text())) + uint(linebreaks)
		start := curr
		curr += lineLen

		if curr <= pos {
			continue
		}

		return line, (pos - start + 1)
	}

	return 0, 0
}

func LocToPos(content string, row uint, col uint) uint {
	scanner := bufio.NewScanner(strings.NewReader(content))

	linebreaks := 1
	if strings.Contains(content, "\r\n") {
		linebreaks = 2
	}

	var line uint
	var curr uint
	for scanner.Scan() {
		line++

		if line < row {
			lineLen := uint(len(scanner.Text())) + uint(linebreaks)
			curr += lineLen
			continue
		}

		return curr + col - 1
	}

	return 0
}
