package position

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/VKCOM/php-parser/pkg/position"
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

func FromIRPosition(path, content string, index int) *Position {
	row, col := PosToLoc(content, uint(index))
	return &Position{
		Row:  row,
		Col:  col,
		Path: path,
	}
}

func (p *Position) String() string {
	return fmt.Sprintf("%s(%d:%d)", p.Path, p.Row, p.Col)
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

func (p *Position) ToIRPosition(content string) *position.Position {
	irPos := int(LocToPos(content, p.Row, p.Col))
	row := int(p.Row)

	return &position.Position{
		StartLine: row,
		EndLine:   row,
		StartPos:  irPos,
		EndPos:    irPos,
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
