package project

import (
	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/VKCOM/php-parser/pkg/visitor"
)

func NewDefinitionTraverser(row int, col int) DefinitionResolver {
	return DefinitionResolver{
		col:   col,
		row:   row,
		Lines: make(map[int]*line),
	}
}

type DefinitionResolver struct {
	visitor.Null
	col int
	row int

	Lines map[int]*line
}

func newLine() *line {
	return &line{
		startPos: -1,
		endPos:   -1,
		nodes:    make([]ast.Vertex, 0),
	}
}

type line struct {
	startPos int
	endPos   int

	nodes []ast.Vertex
}

func (v *DefinitionResolver) getNode() (ast.Vertex, bool) {
	line, exists := v.Lines[v.row]
	if !exists {
		return nil, false
	}

	pos := line.startPos + v.col

	var min ast.Vertex
	for _, node := range line.nodes {
		if node.GetPosition().StartPos <= pos && node.GetPosition().EndPos >= pos {
			if min == nil {
				min = node
				continue
			}

			range1 := node.GetPosition().EndPos - node.GetPosition().StartPos
			range2 := min.GetPosition().EndPos - min.GetPosition().StartPos

			if range1 < range2 {
				min = node
			}
		}
	}

	return min, min != nil
}

func (v *DefinitionResolver) check(n ast.Vertex) {
	pos := n.GetPosition()

	for i := pos.StartLine; i <= pos.EndLine; i++ {
		if _, exists := v.Lines[i]; !exists {
			v.Lines[i] = newLine()
		}

		v.Lines[i].nodes = append(v.Lines[i].nodes, n)
	}

	startLine := v.Lines[pos.StartLine]
	endLine := v.Lines[pos.EndLine]

	// The current node starts earlier, adjust line.
	if startLine.startPos == -1 || startLine.startPos > pos.StartPos {
		startLine.startPos = pos.StartPos
	}

	// The current node end later, adjust line.
	if endLine.endPos < pos.EndPos {
		endLine.endPos = pos.EndPos
	}
}

func (v *DefinitionResolver) Nullable(n *ast.Nullable) {
	v.check(n)
}

func (v *DefinitionResolver) Parameter(n *ast.Parameter) {
	v.check(n)
}

func (v *DefinitionResolver) Identifier(n *ast.Identifier) {
	v.check(n)
}

func (v *DefinitionResolver) Argument(n *ast.Argument) {
	v.check(n)
}

func (v *DefinitionResolver) MatchArm(n *ast.MatchArm) {
	v.check(n)
}

func (v *DefinitionResolver) Union(n *ast.Union) {
	v.check(n)
}

func (v *DefinitionResolver) Intersection(n *ast.Intersection) {
	v.check(n)
}

func (v *DefinitionResolver) Attribute(n *ast.Attribute) {
	v.check(n)
}

func (v *DefinitionResolver) AttributeGroup(n *ast.AttributeGroup) {
	v.check(n)
}

func (v *DefinitionResolver) StmtBreak(n *ast.StmtBreak) {
	v.check(n)
}

func (v *DefinitionResolver) StmtCase(n *ast.StmtCase) {
	v.check(n)
}

func (v *DefinitionResolver) StmtCatch(n *ast.StmtCatch) {
	v.check(n)
}

func (v *DefinitionResolver) StmtEnum(n *ast.StmtEnum) {
	v.check(n)
}

func (v *DefinitionResolver) EnumCase(n *ast.EnumCase) {
	v.check(n)
}

func (v *DefinitionResolver) StmtClass(n *ast.StmtClass) {
	v.check(n)
}

func (v *DefinitionResolver) StmtClassConstList(n *ast.StmtClassConstList) {
	v.check(n)
}

func (v *DefinitionResolver) StmtClassMethod(n *ast.StmtClassMethod) {
	v.check(n)
}

func (v *DefinitionResolver) StmtConstList(n *ast.StmtConstList) {
	v.check(n)
}

func (v *DefinitionResolver) StmtConstant(n *ast.StmtConstant) {
	v.check(n)
}

func (v *DefinitionResolver) StmtContinue(n *ast.StmtContinue) {
	v.check(n)
}

func (v *DefinitionResolver) StmtDeclare(n *ast.StmtDeclare) {
	v.check(n)
}

func (v *DefinitionResolver) StmtDefault(n *ast.StmtDefault) {
	v.check(n)
}

func (v *DefinitionResolver) StmtDo(n *ast.StmtDo) {
	v.check(n)
}

func (v *DefinitionResolver) StmtEcho(n *ast.StmtEcho) {
	v.check(n)
}

func (v *DefinitionResolver) StmtElse(n *ast.StmtElse) {
	v.check(n)
}

func (v *DefinitionResolver) StmtElseIf(n *ast.StmtElseIf) {
	v.check(n)
}

func (v *DefinitionResolver) StmtExpression(n *ast.StmtExpression) {
	v.check(n)
}

func (v *DefinitionResolver) StmtFinally(n *ast.StmtFinally) {
	v.check(n)
}

func (v *DefinitionResolver) StmtFor(n *ast.StmtFor) {
	v.check(n)
}

func (v *DefinitionResolver) StmtForeach(n *ast.StmtForeach) {
	v.check(n)
}

func (v *DefinitionResolver) StmtFunction(n *ast.StmtFunction) {
	v.check(n)
}

func (v *DefinitionResolver) StmtGlobal(n *ast.StmtGlobal) {
	v.check(n)
}

func (v *DefinitionResolver) StmtGoto(n *ast.StmtGoto) {
	v.check(n)
}

func (v *DefinitionResolver) StmtHaltCompiler(n *ast.StmtHaltCompiler) {
	v.check(n)
}

func (v *DefinitionResolver) StmtIf(n *ast.StmtIf) {
	v.check(n)
}

func (v *DefinitionResolver) StmtInlineHtml(n *ast.StmtInlineHtml) {
	v.check(n)
}

func (v *DefinitionResolver) StmtInterface(n *ast.StmtInterface) {
	v.check(n)
}

func (v *DefinitionResolver) StmtLabel(n *ast.StmtLabel) {
	v.check(n)
}

func (v *DefinitionResolver) StmtNamespace(n *ast.StmtNamespace) {
	v.check(n)
}

func (v *DefinitionResolver) StmtNop(n *ast.StmtNop) {
	v.check(n)
}

func (v *DefinitionResolver) StmtProperty(n *ast.StmtProperty) {
	v.check(n)
}

func (v *DefinitionResolver) StmtPropertyList(n *ast.StmtPropertyList) {
	v.check(n)
}

func (v *DefinitionResolver) StmtReturn(n *ast.StmtReturn) {
	v.check(n)
}

func (v *DefinitionResolver) StmtStatic(n *ast.StmtStatic) {
	v.check(n)
}

func (v *DefinitionResolver) StmtStaticVar(n *ast.StmtStaticVar) {
	v.check(n)
}

func (v *DefinitionResolver) StmtStmtList(n *ast.StmtStmtList) {
	v.check(n)
}

func (v *DefinitionResolver) StmtSwitch(n *ast.StmtSwitch) {
	v.check(n)
}

func (v *DefinitionResolver) StmtThrow(n *ast.StmtThrow) {
	v.check(n)
}

func (v *DefinitionResolver) StmtTrait(n *ast.StmtTrait) {
	v.check(n)
}

func (v *DefinitionResolver) StmtTraitUse(n *ast.StmtTraitUse) {
	v.check(n)
}

func (v *DefinitionResolver) StmtTraitUseAlias(n *ast.StmtTraitUseAlias) {
	v.check(n)
}

func (v *DefinitionResolver) StmtTraitUsePrecedence(n *ast.StmtTraitUsePrecedence) {
	v.check(n)
}

func (v *DefinitionResolver) StmtTry(n *ast.StmtTry) {
	v.check(n)
}

func (v *DefinitionResolver) StmtUnset(n *ast.StmtUnset) {
	v.check(n)
}

func (v *DefinitionResolver) StmtUse(n *ast.StmtUseList) {
	v.check(n)
}

func (v *DefinitionResolver) StmtGroupUse(n *ast.StmtGroupUseList) {
	v.check(n)
}

func (v *DefinitionResolver) StmtUseDeclaration(n *ast.StmtUse) {
	v.check(n)
}

func (v *DefinitionResolver) StmtWhile(n *ast.StmtWhile) {
	v.check(n)
}

func (v *DefinitionResolver) ExprArray(n *ast.ExprArray) {
	v.check(n)
}

func (v *DefinitionResolver) ExprArrayDimFetch(n *ast.ExprArrayDimFetch) {
	v.check(n)
}

func (v *DefinitionResolver) ExprArrayItem(n *ast.ExprArrayItem) {
	v.check(n)
}

func (v *DefinitionResolver) ExprArrowFunction(n *ast.ExprArrowFunction) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBrackets(n *ast.ExprBrackets) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBitwiseNot(n *ast.ExprBitwiseNot) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBooleanNot(n *ast.ExprBooleanNot) {
	v.check(n)
}

func (v *DefinitionResolver) ExprClassConstFetch(n *ast.ExprClassConstFetch) {
	v.check(n)
}

func (v *DefinitionResolver) ExprClone(n *ast.ExprClone) {
	v.check(n)
}

func (v *DefinitionResolver) ExprClosure(n *ast.ExprClosure) {
	v.check(n)
}

func (v *DefinitionResolver) ExprClosureUse(n *ast.ExprClosureUse) {
	v.check(n)
}

func (v *DefinitionResolver) ExprConstFetch(n *ast.ExprConstFetch) {
	v.check(n)
}

func (v *DefinitionResolver) ExprEmpty(n *ast.ExprEmpty) {
	v.check(n)
}

func (v *DefinitionResolver) ExprErrorSuppress(n *ast.ExprErrorSuppress) {
	v.check(n)
}

func (v *DefinitionResolver) ExprEval(n *ast.ExprEval) {
	v.check(n)
}

func (v *DefinitionResolver) ExprExit(n *ast.ExprExit) {
	v.check(n)
}

func (v *DefinitionResolver) ExprFunctionCall(n *ast.ExprFunctionCall) {
	v.check(n)
}

func (v *DefinitionResolver) ExprInclude(n *ast.ExprInclude) {
	v.check(n)
}

func (v *DefinitionResolver) ExprIncludeOnce(n *ast.ExprIncludeOnce) {
	v.check(n)
}

func (v *DefinitionResolver) ExprInstanceOf(n *ast.ExprInstanceOf) {
	v.check(n)
}

func (v *DefinitionResolver) ExprIsset(n *ast.ExprIsset) {
	v.check(n)
}

func (v *DefinitionResolver) ExprList(n *ast.ExprList) {
	v.check(n)
}

func (v *DefinitionResolver) ExprMethodCall(n *ast.ExprMethodCall) {
	v.check(n)
}

func (v *DefinitionResolver) ExprNullsafeMethodCall(n *ast.ExprNullsafeMethodCall) {
	v.check(n)
}

func (v *DefinitionResolver) ExprMatch(n *ast.ExprMatch) {
	v.check(n)
}

func (v *DefinitionResolver) ExprNew(n *ast.ExprNew) {
	v.check(n)
}

func (v *DefinitionResolver) ExprPostDec(n *ast.ExprPostDec) {
	v.check(n)
}

func (v *DefinitionResolver) ExprPostInc(n *ast.ExprPostInc) {
	v.check(n)
}

func (v *DefinitionResolver) ExprPreDec(n *ast.ExprPreDec) {
	v.check(n)
}

func (v *DefinitionResolver) ExprPreInc(n *ast.ExprPreInc) {
	v.check(n)
}

func (v *DefinitionResolver) ExprPrint(n *ast.ExprPrint) {
	v.check(n)
}

func (v *DefinitionResolver) ExprPropertyFetch(n *ast.ExprPropertyFetch) {
	v.check(n)
}

func (v *DefinitionResolver) ExprNullsafePropertyFetch(n *ast.ExprNullsafePropertyFetch) {
	v.check(n)
}

func (v *DefinitionResolver) ExprRequire(n *ast.ExprRequire) {
	v.check(n)
}

func (v *DefinitionResolver) ExprRequireOnce(n *ast.ExprRequireOnce) {
	v.check(n)
}

func (v *DefinitionResolver) ExprShellExec(n *ast.ExprShellExec) {
	v.check(n)
}

func (v *DefinitionResolver) ExprStaticCall(n *ast.ExprStaticCall) {
	v.check(n)
}

func (v *DefinitionResolver) ExprStaticPropertyFetch(n *ast.ExprStaticPropertyFetch) {
	v.check(n)
}

func (v *DefinitionResolver) ExprTernary(n *ast.ExprTernary) {
	v.check(n)
}

func (v *DefinitionResolver) ExprThrow(n *ast.ExprThrow) {
	v.check(n)
}

func (v *DefinitionResolver) ExprUnaryMinus(n *ast.ExprUnaryMinus) {
	v.check(n)
}

func (v *DefinitionResolver) ExprUnaryPlus(n *ast.ExprUnaryPlus) {
	v.check(n)
}

func (v *DefinitionResolver) ExprVariable(n *ast.ExprVariable) {
	v.check(n)
}

func (v *DefinitionResolver) ExprYield(n *ast.ExprYield) {
	v.check(n)
}

func (v *DefinitionResolver) ExprYieldFrom(n *ast.ExprYieldFrom) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssign(n *ast.ExprAssign) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignReference(n *ast.ExprAssignReference) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignBitwiseAnd(n *ast.ExprAssignBitwiseAnd) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignBitwiseOr(n *ast.ExprAssignBitwiseOr) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignBitwiseXor(n *ast.ExprAssignBitwiseXor) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignCoalesce(n *ast.ExprAssignCoalesce) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignConcat(n *ast.ExprAssignConcat) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignDiv(n *ast.ExprAssignDiv) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignMinus(n *ast.ExprAssignMinus) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignMod(n *ast.ExprAssignMod) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignMul(n *ast.ExprAssignMul) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignPlus(n *ast.ExprAssignPlus) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignPow(n *ast.ExprAssignPow) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignShiftLeft(n *ast.ExprAssignShiftLeft) {
	v.check(n)
}

func (v *DefinitionResolver) ExprAssignShiftRight(n *ast.ExprAssignShiftRight) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryBitwiseAnd(n *ast.ExprBinaryBitwiseAnd) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryBitwiseOr(n *ast.ExprBinaryBitwiseOr) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryBitwiseXor(n *ast.ExprBinaryBitwiseXor) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryBooleanAnd(n *ast.ExprBinaryBooleanAnd) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryBooleanOr(n *ast.ExprBinaryBooleanOr) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryCoalesce(n *ast.ExprBinaryCoalesce) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryConcat(n *ast.ExprBinaryConcat) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryDiv(n *ast.ExprBinaryDiv) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryEqual(n *ast.ExprBinaryEqual) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryGreater(n *ast.ExprBinaryGreater) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryGreaterOrEqual(n *ast.ExprBinaryGreaterOrEqual) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryIdentical(n *ast.ExprBinaryIdentical) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryLogicalAnd(n *ast.ExprBinaryLogicalAnd) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryLogicalOr(n *ast.ExprBinaryLogicalOr) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryLogicalXor(n *ast.ExprBinaryLogicalXor) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryMinus(n *ast.ExprBinaryMinus) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryMod(n *ast.ExprBinaryMod) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryMul(n *ast.ExprBinaryMul) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryNotEqual(n *ast.ExprBinaryNotEqual) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryNotIdentical(n *ast.ExprBinaryNotIdentical) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryPlus(n *ast.ExprBinaryPlus) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryPow(n *ast.ExprBinaryPow) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryShiftLeft(n *ast.ExprBinaryShiftLeft) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinaryShiftRight(n *ast.ExprBinaryShiftRight) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinarySmaller(n *ast.ExprBinarySmaller) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinarySmallerOrEqual(n *ast.ExprBinarySmallerOrEqual) {
	v.check(n)
}

func (v *DefinitionResolver) ExprBinarySpaceship(n *ast.ExprBinarySpaceship) {
	v.check(n)
}

func (v *DefinitionResolver) ExprCastArray(n *ast.ExprCastArray) {
	v.check(n)
}

func (v *DefinitionResolver) ExprCastBool(n *ast.ExprCastBool) {
	v.check(n)
}

func (v *DefinitionResolver) ExprCastDouble(n *ast.ExprCastDouble) {
	v.check(n)
}

func (v *DefinitionResolver) ExprCastInt(n *ast.ExprCastInt) {
	v.check(n)
}

func (v *DefinitionResolver) ExprCastObject(n *ast.ExprCastObject) {
	v.check(n)
}

func (v *DefinitionResolver) ExprCastString(n *ast.ExprCastString) {
	v.check(n)
}

func (v *DefinitionResolver) ExprCastUnset(n *ast.ExprCastUnset) {
	v.check(n)
}

func (v *DefinitionResolver) ScalarDnumber(n *ast.ScalarDnumber) {
	v.check(n)
}

func (v *DefinitionResolver) ScalarEncapsed(n *ast.ScalarEncapsed) {
	v.check(n)
}

func (v *DefinitionResolver) ScalarEncapsedStringPart(n *ast.ScalarEncapsedStringPart) {
	v.check(n)
}

func (v *DefinitionResolver) ScalarEncapsedStringVar(n *ast.ScalarEncapsedStringVar) {
	v.check(n)
}

func (v *DefinitionResolver) ScalarEncapsedStringBrackets(n *ast.ScalarEncapsedStringBrackets) {
	v.check(n)
}

func (v *DefinitionResolver) ScalarHeredoc(n *ast.ScalarHeredoc) {
	v.check(n)
}

func (v *DefinitionResolver) ScalarLnumber(n *ast.ScalarLnumber) {
	v.check(n)
}

func (v *DefinitionResolver) ScalarMagicConstant(n *ast.ScalarMagicConstant) {
	v.check(n)
}

func (v *DefinitionResolver) ScalarString(n *ast.ScalarString) {
	v.check(n)
}

func (v *DefinitionResolver) NameName(n *ast.Name) {
	v.check(n)
}

func (v *DefinitionResolver) NameFullyQualified(n *ast.NameFullyQualified) {
	v.check(n)
}

func (v *DefinitionResolver) NameRelative(n *ast.NameRelative) {
	v.check(n)
}

func (v *DefinitionResolver) NameNamePart(n *ast.NamePart) {
	v.check(n)
}
