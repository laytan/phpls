package position

import (
	"bufio"
	"fmt"
	"log"
	"strings"

	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/php-parser/pkg/position"
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

func URIToFile(uri string) string {
	return string(uri)[URIFilePrefixLength:]
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
		StartCol:  int(p.Col),
		EndCol:    int(p.Col),
	}
}

func PosToLoc(content string, pos uint) (row uint, col uint) {
	log.Println(
		"DEPRECATED: migrate from this to using the StartCol and EndCol provided by *position.Position",
	)
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

func AstToLspLocation(path string, p *position.Position) protocol.Location {
	return protocol.Location{
		URI: protocol.DocumentURI("file://" + path),
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      uint32(p.StartLine) - 1,
				Character: uint32(p.StartCol),
			},
			End: protocol.Position{
				Line:      uint32(p.EndLine) - 1,
				Character: uint32(p.EndCol),
			},
		},
	}
}

func FromAst(path string, p *position.Position) *Position {
	return &Position{
		Path: path,
		Row:  uint(p.StartLine),
		Col:  uint(p.StartCol) + 1,
	}
}

func AstInAst(scope *position.Position, p *position.Position) bool {
	if scope == nil || p == nil {
		return false
	}

	if scope.EndLine == 0 && scope.EndPos == 0 {
		return p.StartLine >= scope.StartLine && p.StartPos >= scope.StartPos
	}

	return p.StartLine >= scope.StartLine && p.EndLine <= scope.EndLine &&
		p.StartPos >= scope.StartPos &&
		p.EndPos <= scope.EndPos
}

func PosInAst(scope *position.Position, p *Position) bool {
	if scope == nil {
		return false
	}

	if p.Row == uint(scope.StartLine) {
		return p.Col >= uint(scope.StartCol+1)
	}

	if p.Row == uint(scope.EndLine) {
		return p.Col <= uint(scope.EndCol+1)
	}

	return p.Row >= uint(scope.StartLine) && p.Row <= uint(scope.EndLine)
}
