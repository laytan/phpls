// Code generated by "stringer -type TokenType"; DO NOT EDIT.

package token

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Illegal-0]
	_ = x[EOF-1]
	_ = x[PHPStart-2]
	_ = x[PHPEnd-3]
	_ = x[PHPEchoStart-4]
	_ = x[NonPHP-5]
	_ = x[Ident-6]
	_ = x[KeywordsStart-7]
	_ = x[Function-8]
	_ = x[Return-9]
	_ = x[Abstract-10]
	_ = x[And-11]
	_ = x[As-12]
	_ = x[Break-13]
	_ = x[Callable-14]
	_ = x[Case-15]
	_ = x[Catch-16]
	_ = x[Class-17]
	_ = x[Clone-18]
	_ = x[Const-19]
	_ = x[Continue-20]
	_ = x[Default-21]
	_ = x[Die-22]
	_ = x[Do-23]
	_ = x[Echo-24]
	_ = x[Else-25]
	_ = x[ElseIf-26]
	_ = x[EndDeclare-27]
	_ = x[EndFor-28]
	_ = x[EndForEach-29]
	_ = x[EndIf-30]
	_ = x[EndSwitch-31]
	_ = x[EndWhile-32]
	_ = x[Extends-33]
	_ = x[Final-34]
	_ = x[Finally-35]
	_ = x[Fn-36]
	_ = x[For-37]
	_ = x[ForEach-38]
	_ = x[Global-39]
	_ = x[GoTo-40]
	_ = x[If-41]
	_ = x[Implements-42]
	_ = x[Include-43]
	_ = x[IncludeOnce-44]
	_ = x[InstanceOf-45]
	_ = x[InsteadOf-46]
	_ = x[Interface-47]
	_ = x[KVar-48]
	_ = x[Match-49]
	_ = x[Namespace-50]
	_ = x[New-51]
	_ = x[Or-52]
	_ = x[Print-53]
	_ = x[Private-54]
	_ = x[Protected-55]
	_ = x[Public-56]
	_ = x[Readonly-57]
	_ = x[Require-58]
	_ = x[RequireOnce-59]
	_ = x[Static-60]
	_ = x[Switch-61]
	_ = x[Throw-62]
	_ = x[Trait-63]
	_ = x[Try-64]
	_ = x[Use-65]
	_ = x[While-66]
	_ = x[XOR-67]
	_ = x[Yield-68]
	_ = x[YieldFrom-69]
	_ = x[KeywordsEnd-70]
	_ = x[Var-71]
	_ = x[Number-72]
	_ = x[Assign-73]
	_ = x[Plus-74]
	_ = x[Minus-75]
	_ = x[Comma-76]
	_ = x[Semicolon-77]
	_ = x[LParen-78]
	_ = x[RParen-79]
	_ = x[LBrace-80]
	_ = x[RBrace-81]
	_ = x[LBracket-82]
	_ = x[RBracket-83]
	_ = x[ClassAccess-84]
	_ = x[LineComment-85]
	_ = x[SimpleString-86]
	_ = x[StringStart-87]
	_ = x[StringContent-88]
	_ = x[StringEnd-89]
}

const _TokenType_name = "IllegalEOFPHPStartPHPEndPHPEchoStartNonPHPIdentKeywordsStartFunctionReturnAbstractAndAsBreakCallableCaseCatchClassCloneConstContinueDefaultDieDoEchoElseElseIfEndDeclareEndForEndForEachEndIfEndSwitchEndWhileExtendsFinalFinallyFnForForEachGlobalGoToIfImplementsIncludeIncludeOnceInstanceOfInsteadOfInterfaceKVarMatchNamespaceNewOrPrintPrivateProtectedPublicReadonlyRequireRequireOnceStaticSwitchThrowTraitTryUseWhileXORYieldYieldFromKeywordsEndVarNumberAssignPlusMinusCommaSemicolonLParenRParenLBraceRBraceLBracketRBracketClassAccessLineCommentSimpleStringStringStartStringContentStringEnd"

var _TokenType_index = [...]uint16{0, 7, 10, 18, 24, 36, 42, 47, 60, 68, 74, 82, 85, 87, 92, 100, 104, 109, 114, 119, 124, 132, 139, 142, 144, 148, 152, 158, 168, 174, 184, 189, 198, 206, 213, 218, 225, 227, 230, 237, 243, 247, 249, 259, 266, 277, 287, 296, 305, 309, 314, 323, 326, 328, 333, 340, 349, 355, 363, 370, 381, 387, 393, 398, 403, 406, 409, 414, 417, 422, 431, 442, 445, 451, 457, 461, 466, 471, 480, 486, 492, 498, 504, 512, 520, 531, 542, 554, 565, 578, 587}

func (i TokenType) String() string {
	if i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}
