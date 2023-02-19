package visitor

import (
	"fmt"
	"log"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/token"
	"github.com/VKCOM/php-parser/pkg/visitor"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/phpversion"
	"golang.org/x/exp/slices"
)

type LanguageLevelTypeAware struct {
	visitor.Null

	version *phpversion.PHPVersion
	logging bool

	targetter *targetter
}

func NewLanguageLevelTypeAware(
	version *phpversion.PHPVersion,
	logging bool,
) *LanguageLevelTypeAware {
	return &LanguageLevelTypeAware{
		version: version,
		logging: logging,
		targetter: newTargetter([][]byte{
			[]byte("JetBrains"),
			[]byte("PhpStorm"),
			[]byte("Internal"),
			[]byte("LanguageLevelTypeAware"),
		}),
	}
}

func (e *LanguageLevelTypeAware) Root(n *ast.Root) {
	for _, s := range n.Stmts {
		s.Accept(e)
	}
}

func (e *LanguageLevelTypeAware) StmtNamespace(n *ast.StmtNamespace) {
	exit := e.targetter.EnterNamespace(n)
	defer exit()

	for _, s := range n.Stmts {
		s.Accept(e)
	}
}

func (e *LanguageLevelTypeAware) StmtUse(n *ast.StmtUseList) {
	for _, s := range n.Uses {
		s.Accept(e)
	}
}

func (e *LanguageLevelTypeAware) StmtUseDeclaration(n *ast.StmtUse) {
	e.targetter.EnterUse(n)
}

func (e *LanguageLevelTypeAware) StmtClass(n *ast.StmtClass) {
	for _, s := range n.Stmts {
		s.Accept(e)
	}
}

func (e *LanguageLevelTypeAware) StmtInterface(n *ast.StmtInterface) {
	for _, s := range n.Stmts {
		s.Accept(e)
	}
}

func (e *LanguageLevelTypeAware) StmtTrait(n *ast.StmtTrait) {
	for _, s := range n.Stmts {
		s.Accept(e)
	}
}

func (e *LanguageLevelTypeAware) StmtFunction(n *ast.StmtFunction) {
	for i, attrG := range n.AttrGroups {
		typ := e.checkAttrGroup(attrG.(*ast.AttributeGroup))
		if typ == nil {
			continue
		}

		newFf, currDox := e.getCurrDox(
			n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating,
		)
		n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating = newFf

		doc := findOrAppendDoc(currDox, &phpdoxer.NodeReturn{})
		doc.Type = typ

		n.AttrGroups = slices.Delete(n.AttrGroups, i, i+1)
		e.logRemoval(n)

		addDocToFunc(n, currDox)

		n.ColonTkn = nil
		n.ReturnType = nil
	}

	changed := e.checkParams(n.Params)

	var currDox *phpdoxer.Doc
	switch {
	case len(n.AttrGroups) > 0:
		newFf, d := e.getCurrDox(
			n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating,
		)
		currDox = d
		n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating = newFf
	default:
		newFf, d := e.getCurrDox(n.FunctionTkn.FreeFloating)
		currDox = d
		n.FunctionTkn.FreeFloating = newFf
	}

	for _, change := range changed {
		doc := findOrAppendParam(currDox, change.name)
		doc.Type = change.typ
		doc.Name = change.name
	}

	addDocToFunc(n, currDox)
}

func (e *LanguageLevelTypeAware) StmtPropertyList(n *ast.StmtPropertyList) {
	for i, attrG := range n.AttrGroups {
		typ := e.checkAttrGroup(attrG.(*ast.AttributeGroup))
		if typ == nil {
			continue
		}

		newFf, currDox := e.getCurrDox(
			n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating,
		)
		n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating = newFf

		doc := findOrAppendDoc(currDox, &phpdoxer.NodeVar{})
		doc.Type = typ

		n.AttrGroups = slices.Delete(n.AttrGroups, i, i+1)
		e.logRemoval(n)

		addDocToProp(n, currDox)

		n.Type = nil
	}
}

func (e *LanguageLevelTypeAware) StmtClassMethod(n *ast.StmtClassMethod) {
	for i, attrG := range n.AttrGroups {
		typ := e.checkAttrGroup(attrG.(*ast.AttributeGroup))
		if typ == nil {
			continue
		}

		newFf, currDox := e.getCurrDox(
			n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating,
		)
		n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating = newFf

		doc := findOrAppendDoc(currDox, &phpdoxer.NodeReturn{})
		doc.Type = typ

		n.AttrGroups = slices.Delete(n.AttrGroups, i, i+1)
		e.logRemoval(n)

		addDocToMeth(n, currDox)

		n.ColonTkn = nil
		n.ReturnType = nil
	}

	changed := e.checkParams(n.Params)

	var currDox *phpdoxer.Doc
	switch {
	case len(n.AttrGroups) > 0:
		newFf, d := e.getCurrDox(
			n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating,
		)
		currDox = d
		n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating = newFf
	case len(n.Modifiers) > 0:
		newFf, d := e.getCurrDox(n.Modifiers[0].(*ast.Identifier).IdentifierTkn.FreeFloating)
		currDox = d
		n.Modifiers[0].(*ast.Identifier).IdentifierTkn.FreeFloating = newFf
	default:
		newFf, d := e.getCurrDox(n.FunctionTkn.FreeFloating)
		currDox = d
		n.FunctionTkn.FreeFloating = newFf
	}

	for _, change := range changed {
		doc := findOrAppendParam(currDox, change.name)
		doc.Type = change.typ
		doc.Name = change.name
	}

	addDocToMeth(n, currDox)
}

func (e *LanguageLevelTypeAware) checkAttrGroup(n *ast.AttributeGroup) phpdoxer.Type {
	attr := n.Attrs[0].(*ast.Attribute)

	if !e.targetter.MatchName(attr.Name) {
		return nil
	}

	if len(n.Attrs) > 1 {
		panic(
			"unexpected LanguageLevelTypeAware attribute in an attributeGroup with > 1 attributes, this case is not handled by this visitor because it wasn't needed before",
		)
	}

	typ := attr.Args[1].(*ast.Argument).Expr.(*ast.ScalarString).Value

	arg1ArrItems := attr.Args[0].(*ast.Argument).Expr.(*ast.ExprArray).Items
	for _, item := range arg1ArrItems {
		// I think this happens when you add an extra comma at the end, not sure.
		// But it results in an extra item with everything set to nil.
		if item.(*ast.ExprArrayItem).Key == nil {
			continue
		}

		version := string(item.(*ast.ExprArrayItem).Key.(*ast.ScalarString).Value)
		version = version[1 : len(version)-1]
		versionO, ok := phpversion.FromString(version)
		if !ok {
			log.Panicf("version %s is not able to be converted into a PHPVersion type", version)
		}

		if e.version.Equals(versionO) || e.version.IsHigherThan(versionO) {
			typ = item.(*ast.ExprArrayItem).Val.(*ast.ScalarString).Value
		}
	}

	typeStr := string(typ)
	typeStr = typeStr[1 : len(typeStr)-1]
	if typeStr == "" {
		typeStr = "mixed"
	}

	typeO, err := phpdoxer.ParseType(typeStr)
	if err != nil {
		panic(err)
	}

	return typeO
}

type paramChange struct {
	name string
	typ  phpdoxer.Type
}

func (e *LanguageLevelTypeAware) checkParams(ns []ast.Vertex) (changed []*paramChange) {
	for _, param := range ns {
		typedParam := param.(*ast.Parameter)

		for j, attrG := range typedParam.AttrGroups {
			typ := e.checkAttrGroup(attrG.(*ast.AttributeGroup))
			if typ == nil {
				continue
			}

			changed = append(changed, &paramChange{
				name: string(typedParam.Var.(*ast.ExprVariable).Name.(*ast.Identifier).Value),
				typ:  typ,
			})

			typedParam.AttrGroups = slices.Delete(typedParam.AttrGroups, j, j+1)
			e.logRemoval(param)

			typedParam.Type = nil
		}
	}

	return changed
}

func (e *LanguageLevelTypeAware) getCurrDox(n []*token.Token) ([]*token.Token, *phpdoxer.Doc) {
	for i, c := range n {
		if c.ID != token.T_DOC_COMMENT {
			continue
		}

		doc, err := phpdoxer.ParseFullDoc(string(c.Value))
		if err != nil {
			panic(err)
		}

		n = slices.Delete(n, i, i+1)
		return n, doc
	}

	return n, &phpdoxer.Doc{
		Nodes: []phpdoxer.Node{},
	}
}

// Finds or adds the node to the doc.
// Returns the found/added doc.
func findOrAppendDoc[T phpdoxer.Node](currDox *phpdoxer.Doc, def T) T {
	for _, dn := range currDox.Nodes {
		if typedDn, ok := dn.(T); ok {
			return typedDn
		}
	}

	currDox.Nodes = append(currDox.Nodes, def)
	return def
}

func findOrAppendParam(currDox *phpdoxer.Doc, name string) *phpdoxer.NodeParam {
	for _, dn := range currDox.Nodes {
		if typedDn, ok := dn.(*phpdoxer.NodeParam); ok {
			if typedDn.Name == name {
				return typedDn
			}
		}
	}

	def := &phpdoxer.NodeParam{}
	currDox.Nodes = append(currDox.Nodes, def)
	return def
}

func addDocToFunc(n *ast.StmtFunction, currDox *phpdoxer.Doc) {
	token := &token.Token{
		ID:    token.T_DOC_COMMENT,
		Value: []byte(currDox.String()),
	}

	switch {
	case len(n.AttrGroups) > 0:
		n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating = append(
			n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating,
			token,
		)
	default:
		n.FunctionTkn.FreeFloating = append(n.FunctionTkn.FreeFloating, token)
	}
}

func addDocToMeth(n *ast.StmtClassMethod, currDox *phpdoxer.Doc) {
	token := &token.Token{
		ID:    token.T_DOC_COMMENT,
		Value: []byte(currDox.String()),
	}

	switch {
	case len(n.AttrGroups) > 0:
		n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating = append(
			n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating,
			token,
		)
	case len(n.Modifiers) > 0:
		n.Modifiers[0].(*ast.Identifier).IdentifierTkn.FreeFloating = append(
			n.Modifiers[0].(*ast.Identifier).IdentifierTkn.FreeFloating,
			token,
		)
	default:
		n.FunctionTkn.FreeFloating = append(n.FunctionTkn.FreeFloating, token)
	}
}

func addDocToProp(n *ast.StmtPropertyList, currDox *phpdoxer.Doc) {
	token := &token.Token{
		ID:    token.T_DOC_COMMENT,
		Value: []byte(currDox.String()),
	}

	switch {
	case len(n.AttrGroups) > 0:
		n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating = append(
			n.AttrGroups[0].(*ast.AttributeGroup).OpenAttributeTkn.FreeFloating,
			token,
		)
	default:
		n.Modifiers[0].(*ast.Identifier).IdentifierTkn.FreeFloating = append(
			n.Modifiers[0].(*ast.Identifier).IdentifierTkn.FreeFloating,
			token,
		)
	}
}

// TODO: don't require arg.
func (e *LanguageLevelTypeAware) logRemoval(n ast.Vertex) {
	if e.logging {
		_, _ = fmt.Printf("x") //nolint:forbidigo // For visualization.
	}
}
