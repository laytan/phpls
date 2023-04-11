package project

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/nodeident"
	"github.com/laytan/elephp/pkg/position"
	"github.com/laytan/elephp/pkg/traversers"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/php-parser/pkg/token"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
)

// TODO: test out returning all results
const maxCompletionResults = 20

var ErrNoCompletionResults = errors.New("No completion results found for symbol at given position")

func (p *Project) Complete(
	pos *position.Position,
) []*index.INode {
	query := p.getCompletionQuery(pos)
	if query == "" {
		return nil
	}

	return index.Current.FindPrefix(query, maxCompletionResults)
}

//go:generate stringer -type=CompletionAction
type CompletionAction int

const (
	None               CompletionAction = iota // Not doing anything special, should complete global stuff, available variables in the scope.
	NamingFunction                             // A function is being named, probably nothing to complete.
	NamingClassLike                            // A class, interface or trait is being named, probably nothing to complete.
	NamingNamespace                            // Typing after 'namespace', complete with possible namespaces based on directory structure.
	StartingVariable                           // A '$' has been typed, should complete available variables in scope.
	Variable                                   // A '$' with some characters (variable being typed), should complete available variables in scope, starting with the currently typed prefix.
	ObjectOp                                   // Last op was a '->', should resolve up to before '->' and complete available stuff on that variable.
	StaticOp                                   // Last op was '::', should resolve up to before '::' and complete available static stuff on that variable, can also complete 'class' to make $obj::class.
	Implements                                 // Typing after 'implements', should complete any interfaces.
	Extends                                    // Typing after 'extends', should complete any classes.
	Use                                        // Typing after 'use', should complete any traits.
	AddingFile                                 // Typing after include, include once, require or require once. Should complete file paths.
	Newing                                     // Typing after 'new'. Should complete classes.
	Name                                       // Typing a bare name, complete any classes that match it.
	NameRelative                               // Typing 'namespace\foo'. Should complete classes relative to current namespace.
	NameFullyQualified                         // Typing a name starting with '\'. Should complete any FQN starting with what has been typed.
)

type CompletionContext struct {
	Action CompletionAction
	Tokens []*token.Token // The last of these is the place the cursor is at.
}

func GetCompletionQuery2(pos *position.Position) CompletionContext {
	ctx := CompletionContext{}
	lexer := wrkspc.Current.FLexerOf(pos.Path)
	for tok := lexer.Lex(); tok != nil && tok.ID != 0; tok = lexer.Lex() {
		if tok.Position.EndLine > int(pos.Row) {
			break
		}

		if tok.Position.EndLine == int(pos.Row) {
			if tok.Position.StartCol > int(pos.Col) {
				break
			}
		}

		skipAdd := false
		switch tok.ID { //nolint:exhaustive // Does not need to be exhaustive.
		case token.T_INCLUDE, token.T_INCLUDE_ONCE, token.T_REQUIRE, token.T_REQUIRE_ONCE:
			ctx.Action = AddingFile
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_NEW:
			ctx.Action = Newing
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_NAME_QUALIFIED:
			ctx.Action = Name
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_NAME_FULLY_QUALIFIED, token.T_NS_SEPARATOR:
			ctx.Action = NameFullyQualified
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_NAME_RELATIVE:
			ctx.Action = NameRelative
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_NAMESPACE:
			next := lexer.Lex()
			if next.ID == token.T_NS_SEPARATOR { // Handle 'namespace\' not being a T_NAME_RELATIVE yet.
				ctx.Action = NameRelative
				ctx.Tokens = ctx.Tokens[:0]
				ctx.Tokens = append(ctx.Tokens, tok, next)
				skipAdd = true
				break
			}

			ctx.Action = NamingNamespace
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_FUNCTION:
			ctx.Action = NamingFunction
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_CLASS:
			if ctx.Action == StaticOp { // Prevent ::class from going here.
				break
			}

			ctx.Action = NamingClassLike
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_TRAIT, token.T_INTERFACE:
			ctx.Action = NamingClassLike
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_IMPLEMENTS:
			ctx.Action = Implements
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_EXTENDS:
			ctx.Action = Extends
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_USE:
			ctx.Action = Use
			ctx.Tokens = ctx.Tokens[:0]
		case 36: // a lone '$'
			ctx.Action = StartingVariable
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_VARIABLE:
			ctx.Action = Variable
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_OBJECT_OPERATOR:
			ctx.Action = ObjectOp
		case token.T_PAAMAYIM_NEKUDOTAYIM:
			ctx.Action = StaticOp
		case 59: // a lone ';'
			ctx.Action = None
			ctx.Tokens = ctx.Tokens[:0]
			skipAdd = true
		}

		if !skipAdd {
			ctx.Tokens = append(ctx.Tokens, tok)
		}
	}

	return ctx
}

func INodeCompletionKind(kind ast.Type) protocol.CompletionItemKind {
	switch kind {
	case ast.TypeStmtFunction:
		return protocol.FunctionCompletion
	case ast.TypeStmtClass, ast.TypeStmtTrait: // Trait doesn't really have a matching kind.
		return protocol.ClassCompletion
	case ast.TypeStmtInterface:
		return protocol.InterfaceCompletion
	default:
		return 0
	}
}

type CompletionItemData struct {
	CurrPos   *position.Position
	TargetPos *position.Position
}

func CompletionData(pos *position.Position, n *index.INode) string {
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

func Complete(pos *position.Position, comp CompletionContext) (list protocol.CompletionList) {
	switch comp.Action {
	case None:
		list.Items = functional.Map(Keywords, func(keyword string) protocol.CompletionItem {
			return protocol.CompletionItem{Label: keyword, Kind: protocol.KeywordCompletion}
		})
	case Name:
		query := string(comp.Tokens[len(comp.Tokens)-1].Value)
		log.Printf("[INFO]: looking for classes starting with: %s", query)
		indexNodes := index.Current.FindPrefix(
			query,
			maxCompletionResults,
			ast.TypeStmtClass,
			ast.TypeStmtInterface,
			ast.TypeStmtTrait,
		)

		list.Items = functional.Map(indexNodes, func(node *index.INode) protocol.CompletionItem {
			return protocol.CompletionItem{
				Label: node.Identifier,
				Kind:  INodeCompletionKind(node.Kind),
				Data:  CompletionData(pos, node),
				// Adding an additional text edit, so the client shows it in the UI.
				// The actual text edit is added in the Resolve method.
				AdditionalTextEdits: []protocol.TextEdit{{}},
			}
		})
	case NameFullyQualified:
		query := string(comp.Tokens[len(comp.Tokens)-1].Value)
		log.Printf("[INFO]: looking for classes starting with: %s", query)
		indexNodes := index.Current.FindFqnPrefix(
			query,
			maxCompletionResults,
			ast.TypeStmtClass,
			ast.TypeStmtInterface,
			ast.TypeStmtTrait,
		)
		list.Items = functional.Map(indexNodes, func(node *index.INode) protocol.CompletionItem {
			return protocol.CompletionItem{
				Label: node.FQN.String(),
				Kind:  INodeCompletionKind(node.Kind),
			}
		})
	case NameRelative:
		root := wrkspc.Current.FIROf(pos.Path)
		v := traversers.NewNamespace(int(pos.Row))
		tv := traverser.NewTraverser(v)
		root.Accept(tv)
		namespace := nodeident.Get(v.Result)

		query := string(comp.Tokens[len(comp.Tokens)-1].Value)
		noNsQuery := strings.TrimPrefix(query, "namespace")
		fullQuery := namespace + noNsQuery
		log.Printf("[INFO]: looking for classes starting with: %s", fullQuery)
		indexNodes := index.Current.FindFqnPrefix(
			fullQuery,
			maxCompletionResults,
			ast.TypeStmtClass,
			ast.TypeStmtInterface,
			ast.TypeStmtTrait,
		)
		list.Items = functional.Map(indexNodes, func(node *index.INode) protocol.CompletionItem {
			label := strings.Replace(node.FQN.String(), namespace, "namespace", 1)
			return protocol.CompletionItem{
				Label: label,
				Kind:  INodeCompletionKind(node.Kind),
			}
		})
	case NamingNamespace:
		if prediction := PredictNamespace(pos); prediction != "" {
			list.Items = []protocol.CompletionItem{{
				Label: prediction,
				Kind:  protocol.ModuleCompletion,
			}}
		}
	default:
	}

	list.IsIncomplete = true
	return list
}

// 1. Check if there are other sibling files, parse for their namespace.
// 2. Keep going up a dir, trying the files for their namespace, camel case the directory names.
// 3. TODO: If we find the root of the project, return.
func PredictNamespace(pos *position.Position) string {
	var dir, name string
	dir = pos.Path

	suffix := ""
	for i := 0; dir != ""; i++ {
		dir, name = filepath.Split(dir)
		if len(dir) > 0 {
			dir = dir[:len(dir)-1]
		}

		if i != 0 {
			n := strings.ToUpper(name[:1]) + name[1:]
			suffix = "\\" + n + suffix
		}

		files, err := os.ReadDir(dir)
		if err != nil {
			log.Println(err)
			return ""
		}

		for _, file := range files {
			fn := filepath.Join(dir, file.Name())
			if pos.Path == fn {
				continue
			}

			if !wrkspc.Current.IsPhpFile(fn) {
				continue
			}

			root, err := wrkspc.Current.IROf(fn)
			if err != nil {
				log.Println(err)
				continue
			}

			v := traversers.NewNamespaceFirstResult()
			tv := traverser.NewTraverser(v)
			root.Accept(tv)
			if v.Result == nil {
				log.Printf("[DEBUG]: %q has no namespace", file.Name())
				continue
			}

			return nodeident.Get(v.Result)[1:] + suffix
		}
	}

	return ""
}

// Gets the current word ([a-zA-Z0-9]*) that the position is at.
func (p *Project) getCompletionQuery(pos *position.Position) string {
	content := wrkspc.Current.FContentOf(pos.Path)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for i := 0; scanner.Scan(); i++ {
		// The target line:
		if uint(i) != pos.Row-1 {
			continue
		}

		content := scanner.Text()
		rContent := []rune(content)
		if len(rContent) == 0 {
			break
		}

		start := uint(0)
		end := uint(len(rContent))
		startI := pos.Col - 2
		if startI >= end {
			startI = end - 1
		}

		for i := startI; i > 0; i-- {
			ch := rContent[i]

			if unicode.IsDigit(ch) || unicode.IsLetter(ch) {
				continue
			}

			start = i + 1
			break
		}

		for i := startI; int(i) < len(rContent); i++ {
			ch := rContent[i]

			if unicode.IsDigit(ch) || unicode.IsLetter(ch) {
				continue
			}

			end = i
			break
		}

		if start >= end {
			return ""
		}

		return strings.TrimSpace(string(rContent[start:end]))
	}

	return ""
}
