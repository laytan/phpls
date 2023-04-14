package project

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/laytan/phpls/internal/context"
	"github.com/laytan/phpls/internal/expr"
	"github.com/laytan/phpls/internal/index"
	"github.com/laytan/phpls/internal/project/definition"
	"github.com/laytan/phpls/internal/symbol"
	"github.com/laytan/phpls/internal/wrkspc"
	"github.com/laytan/phpls/pkg/nodeident"
	"github.com/laytan/phpls/pkg/nodescopes"
	"github.com/laytan/phpls/pkg/parsing"
	"github.com/laytan/phpls/pkg/phpversion"
	"github.com/laytan/phpls/pkg/position"
	"github.com/laytan/phpls/pkg/set"
	"github.com/laytan/phpls/pkg/traversers"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/php-parser/pkg/ast"
	astposition "github.com/laytan/php-parser/pkg/position"
	"github.com/laytan/php-parser/pkg/token"
	"github.com/laytan/php-parser/pkg/visitor"
	"github.com/laytan/php-parser/pkg/visitor/traverser"
)

// TODO: test out returning all results.
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

func GetCompletionQuery(pos *position.Position) CompletionContext {
	// content, root := wrkspc.Current.FAllOf(pos.Path)
	// offset := position.LocToPos(content, pos.Row, pos.Col)
	// v := traversers.NewNodeAtPos(int(offset))
	// tv := traverser.NewTraverser(v)
	// root.Accept(tv)
	// what.Is(v.Nodes)

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
		case 36: // '$'
			ctx.Action = StartingVariable
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_VARIABLE:
			ctx.Action = Variable
			ctx.Tokens = ctx.Tokens[:0]
		case token.T_OBJECT_OPERATOR:
			ctx.Action = ObjectOp
		case token.T_PAAMAYIM_NEKUDOTAYIM:
			ctx.Action = StaticOp
		case 59, 123, 125: // ';', '}', '}'
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

func NodeCompletionKind(kind ast.Type) protocol.CompletionItemKind {
	switch kind {
	case ast.TypeStmtFunction:
		return protocol.FunctionCompletion
	case ast.TypeStmtClass, ast.TypeStmtTrait: // Trait doesn't really have a matching kind.
		return protocol.ClassCompletion
	case ast.TypeStmtInterface:
		return protocol.InterfaceCompletion
	case ast.TypeStmtClassMethod:
		return protocol.MethodCompletion
	case ast.TypeParameter: // Parameter is a property, because this is only called with constructor promoted properties.
		return protocol.PropertyCompletion
	case ast.TypeStmtPropertyList:
		return protocol.PropertyCompletion
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
	var lastValue string
	if len(comp.Tokens) > 0 {
		lastValue = string(comp.Tokens[len(comp.Tokens)-1].Value)
	}

	var ctx *context.Ctx
	needCtx := func() bool {
		if ctx != nil {
			return true
		}

		c, err := context.New(pos)
		if err != nil {
			log.Printf("unable to create context at %v: %v", pos, err)
			return false
		}

		ctx = c
		return true
	}

	switch comp.Action {
	case NamingFunction, NamingClassLike:
		break
	case None:
		AddKeywordsWithPrefix(&list, lastValue)

		// Lets not return every single result in our index.
		if lastValue != "" {
			AddFromIndex(&list, pos, lastValue)
		}

		// Complete variables when no query,
		// If user has typed something and still in None, they don't want variables.
		// Because the first character is always a '$'.
		if lastValue == "" {
			if !needCtx() {
				break
			}
			AddScopeVars(&list, ctx, lastValue)
		}
	case StartingVariable, Variable:
		if !needCtx() {
			break
		}
		AddScopeVars(&list, ctx, lastValue)
	case Implements:
		AddFromIndex(&list, pos, lastValue, ast.TypeStmtInterface)
	case Use:
		AddFromIndex(&list, pos, lastValue, ast.TypeStmtTrait)
	case Extends, Newing:
		AddFromIndex(&list, pos, lastValue, ast.TypeStmtClass)
	case Name:
		AddFromIndex(
			&list,
			pos,
			lastValue,
			ast.TypeStmtClass,
			ast.TypeStmtInterface,
			ast.TypeStmtTrait,
		)
	case ObjectOp:
		if len(comp.Tokens) == 1 {
			break
		}

		if !needCtx() {
			break
		}

		AddExprCompletions(&list, comp, ctx, lastValue, symbol.FilterNotStatic[symbol.Member]())
	case StaticOp:
		if len(comp.Tokens) == 1 {
			break
		}

		if !needCtx() {
			break
		}

		AddExprCompletions(&list, comp, ctx, lastValue, symbol.FilterStatic[symbol.Member]())
	case NameFullyQualified:
		indexNodes := index.Current.FindFqnPrefix(
			lastValue,
			maxCompletionResults,
			ast.TypeStmtClass,
			ast.TypeStmtInterface,
			ast.TypeStmtTrait,
		)
		for _, node := range indexNodes {
			list.Items = append(list.Items, protocol.CompletionItem{
				Label: node.FQN.String(),
				Kind:  NodeCompletionKind(node.Kind),
			})
		}
	case NameRelative:
		root := wrkspc.Current.FIROf(pos.Path)
		v := traversers.NewNamespace(int(pos.Row))
		tv := traverser.NewTraverser(v)
		root.Accept(tv)
		namespace := nodeident.Get(v.Result)

		noNsQuery := strings.TrimPrefix(lastValue, "namespace")
		fullQuery := namespace + noNsQuery
		indexNodes := index.Current.FindFqnPrefix(
			fullQuery,
			maxCompletionResults,
			ast.TypeStmtClass,
			ast.TypeStmtInterface,
			ast.TypeStmtTrait,
		)
		for _, node := range indexNodes {
			list.Items = append(list.Items, protocol.CompletionItem{
				Label: strings.Replace(node.FQN.String(), namespace, "namespace", 1),
				Kind:  NodeCompletionKind(node.Kind),
			})
		}
	case NamingNamespace:
		if prediction := PredictNamespace(pos); prediction != "" {
			list.Items = append(list.Items, protocol.CompletionItem{
				Label: prediction,
				Kind:  protocol.ModuleCompletion,
			})
		}
	case AddingFile: // NOTE: this is kinda yanky.
		// If 1 token, it is the keyword that started this, don't want that.
		if len(comp.Tokens) == 1 {
			lastValue = ""
		}

		lastValue = strings.Trim(lastValue, "'")
		lastValue = strings.Trim(lastValue, "\"")

		dir := filepath.Dir(pos.Path)
		_, fn := filepath.Split(pos.Path)
		currAbsPath := filepath.Clean(filepath.Join(dir, lastValue))
		err := filepath.WalkDir(currAbsPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			rel, err := filepath.Rel(dir, path)
			if err != nil {
				log.Printf("[ERROR]: can't make relative of %q with base %q: %v", path, dir, err)
				return nil
			}

			if d.IsDir() {
				list.Items = append(
					list.Items,
					protocol.CompletionItem{
						Label: rel + string(filepath.Separator),
						Kind:  protocol.FileCompletion,
					},
				)
				return nil
			}

			if !wrkspc.Current.IsPhpFile(path) {
				return nil
			}

			// Skip current file.
			if fn == rel {
				return nil
			}

			list.Items = append(list.Items, protocol.CompletionItem{
				Label: rel,
				Kind:  protocol.FileCompletion,
			})

			return nil
		})
		if err != nil {
			log.Printf("[ERROR]: walking dir %q: %v", currAbsPath, err)
		}
	}
	return list
}

func AddExprCompletions(
	list *protocol.CompletionList,
	comp CompletionContext,
	ctx *context.Ctx,
	lastValue string,
	additionalFilters ...symbol.FilterFunc[symbol.Member],
) {
	subj := Reconstruct(comp.Tokens[:len(comp.Tokens)-1])
	if subj == nil {
		log.Println(
			"[WARN]: could not reconstruct ast from lexed tokens to provide completions",
		)
		return
	}

	scopes := definition.ContextToScopes(ctx)
	res, lastClass, left := expr.Resolve(subj, scopes)
	if left != 0 {
		log.Println("[WARN]: could not resolve reconstructed expression to provide completion")
		return
	}

	log.Printf("[DEBUG]: res: %#v", res)
	log.Printf("[DEBUG]: lastClass: %#v", lastClass)

	cls, err := symbol.NewClassLikeFromFQN(wrkspc.NewRooter(ctx.Path(), ctx.Root()), lastClass)
	if err != nil {
		log.Printf(
			"[ERROR]: could not create internal class representation for %q: %v",
			lastClass,
			err,
		)
		return
	}

	// TODO: filter privacy.
	members := cls.FindMembersInherit(
		false, // don't shortCircuit.
		symbol.FilterPrefix[symbol.Member](lastValue),
	)
	for _, m := range members {
		item := protocol.CompletionItem{
			Label: strings.TrimPrefix(m.Name(), "$"),
			Kind:  NodeCompletionKind(m.Vertex().GetType()),
		}

		switch tm := m.(type) {
		case *symbol.Method:
			rtype, _, err := tm.Returns()
			if err != nil {
				if !errors.Is(err, symbol.ErrNoReturn) {
					log.Printf("[ERROR]: method %q: %v", tm.Name(), err)
				}
				item.Detail = "mixed"
				break
			}
			item.Detail = rtype.String()
		case *symbol.Property:
			ptype, _, err := tm.Type()
			if err != nil {
				if !errors.Is(err, symbol.ErrNoPropertyType) {
					log.Printf("[ERROR]: method %q: %v", tm.Name(), err)
				}
				item.Detail = "mixed"
				break
			}
			item.Detail = ptype.String()
		}

		list.Items = append(list.Items, item)
	}
}

// Reconstruct tries to construct an AST node from the given tokens.
// It closes any open constructs and tries to turn the tokens into parse-able PHP.
// That PHP is then parsed and the resulting
// nodes have their position set to that of the tokens.
func Reconstruct(tokens []*token.Token) ast.Vertex {
	origB := strings.Builder{}
	for _, t := range tokens {
		_, _ = origB.Write(t.Value)
	}
	orig := origB.String()

	counts := map[rune]int{}
	for _, ch := range orig {
		switch ch {
		case '\'', '"', '(', ')', '[', ']':
			counts[ch]++
		}
	}

	sourceB := strings.Builder{}
	_, _ = sourceB.WriteString(orig)

	if counts['"']%2 != 0 {
		_, _ = sourceB.WriteRune('"')
	}

	if counts['\'']%2 != 0 {
		_, _ = sourceB.WriteRune('\'')
	}

	obc := counts['[']
	cbc := counts[']']
	if obc > cbc {
		for i := cbc; i < obc; i++ {
			_, _ = sourceB.WriteRune(']')
		}
	}

	opc := counts['(']
	cpc := counts[')']
	if opc > cpc {
		for i := cpc; i < opc; i++ {
			_, _ = sourceB.WriteRune(')')
		}
	}

	source := "<?php " + sourceB.String()

	source = strings.TrimSuffix(source, "::")
	source = strings.TrimSuffix(source, "->")

	if !strings.HasSuffix(source, ";") {
		source += ";"
	}

	log.Printf("[INFO]: Reconstruction: %q => %q", orig, source)

	p := parsing.New(phpversion.Latest())
	root, err := p.Parse([]byte(source))
	if err != nil {
		log.Printf("[ERROR]: could not parse %q: %v", source, err)
		return nil
	}

	if len(root.Stmts) == 0 {
		log.Printf("[WARN]: no statements were reconstructed from %q", source)
		return nil
	}

	var res ast.Vertex
	switch tn := root.Stmts[0].(type) {
	case *ast.StmtExpression:
		res = tn.Expr
	default:
		log.Printf("[WARN]: Unexpected reconstruction result of %q: %#v", source, root.Stmts[0])
	}

	if res == nil {
		return nil
	}

	// Traverse to set the position of all nodes to that of the first token.
	// Should allow scope/location based funcs to still work.
	v := &setPosVisitor{Pos: tokens[0].Position}
	tv := traverser.NewTraverser(v)
	res.Accept(tv)

	return res
}

// Sets the position of all nodes in the tree to Pos.
type setPosVisitor struct {
	visitor.Null
	Pos *astposition.Position
}

func (s *setPosVisitor) EnterNode(n ast.Vertex) bool {
	p := n.GetPosition()
	if p == nil {
		return true
	}

	p.StartCol = s.Pos.StartCol
	p.EndCol = s.Pos.EndCol
	p.StartLine = s.Pos.StartLine
	p.EndLine = s.Pos.EndLine
	p.StartPos = s.Pos.StartPos
	p.EndPos = s.Pos.EndPos

	return true
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

func AddFromIndex(
	list *protocol.CompletionList,
	pos *position.Position,
	prefix string,
	types ...ast.Type,
) {
	list.IsIncomplete = true // TODO: set this based on if there are more items with this prefix.
	indexNodes := index.Current.FindPrefix(prefix, maxCompletionResults, types...)
	for _, node := range indexNodes {
		item := protocol.CompletionItem{
			Label: node.Identifier,
			Kind:  NodeCompletionKind(node.Kind),
		}

		// Add completion data for automatic use statements.
		if nodescopes.IsClassLike(node.Kind) {
			item.Data = CompletionData(pos, node)
			// Adding an additional text edit, so the client shows it in the UI.
			// The actual text edit is added in the Resolve method.
			item.AdditionalTextEdits = []protocol.TextEdit{{}}
		}

		list.Items = append(list.Items, item)
	}
}

func AddScopeVars(
	list *protocol.CompletionList,
	ctx *context.Ctx,
	prefix string,
) {
	// TODO: if scope is arrow function, use the scope above that one (vars are inherited in arrow funcs).
	scope := ctx.Scope()
	scopet := scope.GetType()
	if scopet != ast.TypeStmtFunction && scopet != ast.TypeStmtClassMethod &&
		scopet != ast.TypeExprClosure && scopet != ast.TypeExprArrowFunction {
		return
	}

	v := &availableInFuncVisitor{
		Results: set.New[string](),
		Until: &position.Position{
			Row: ctx.Start().Row,
			// Minus length of query, so the current var is not shown, fixes some edge case where '$\n\n$test = '';' would say the $test variable starts at the '$' character.
			Col: ctx.Start().Col - uint(len(prefix)) - 2,
		},
	}
	tv := traverser.NewTraverser(v)
	scope.Accept(tv)

	if scopet == ast.TypeStmtClassMethod {
		v.Results.Add("$this")
	}

	for _, v := range v.Results.Slice() {
		if strings.HasPrefix(v, prefix) {
			list.Items = append(list.Items, protocol.CompletionItem{
				Label: v,
				Kind:  protocol.VariableCompletion,
			})
		}
	}
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

type availableInFuncVisitor struct {
	visitor.Null
	Results *set.Set[string]
	Until   *position.Position
}

var _ ast.Visitor = &availableInFuncVisitor{}

func (a *availableInFuncVisitor) EnterNode(n ast.Vertex) bool {
	p := n.GetPosition()
	if p == nil {
		return true
	}

	// Later line.
	if p.StartLine > int(a.Until.Row) {
		return false
	}

	// Same line, later column.
	if p.StartLine == int(a.Until.Row) && p.StartCol > int(a.Until.Col) {
		return false
	}

	return true
}

func (a *availableInFuncVisitor) ExprVariable(n *ast.ExprVariable) {
	a.Results.Add(nodeident.Get(n))
}
