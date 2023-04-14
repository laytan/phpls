package project

import (
	"strings"

	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/laytan/phpls/pkg/functional"
)

// Source: https://www.php.net/manual/en/reserved.keywords.php
var Keywords = []string{
	"abstract",
	"and",
	"as",
	"break",
	"callable",
	"case",
	"catch",
	"class",
	"clone",
	"const",
	"continue",
	"declare",
	"default",
	"die",
	"do",
	"echo",
	"else",
	"elseif",
	"enddeclare",
	"endfor",
	"endforeach",
	"endif",
	"endswitch",
	"endwhile",
	"exit",
	"extends",
	"final",
	"finally",
	"fn",
	"for",
	"foreach",
	"function",
	"global",
	"goto",
	"if",
	"implements",
	"include",
	"include_once",
	"instanceof",
	"insteadof",
	"interface",
	"match",
	"namespace",
	"new",
	"or",
	"print",
	"private",
	"protected",
	"public",
	"readonly",
	"require",
	"require_once",
	"return",
	"static",
	"switch",
	"throw",
	"trait",
	"try",
	"use",
	"var",
	"while",
	"xor",
	"yield",
	"yield from",
}

var KeywordItems = functional.Map(Keywords, func(keyword string) protocol.CompletionItem {
	return protocol.CompletionItem{Label: keyword, Kind: protocol.KeywordCompletion}
})

func AddKeywordsWithPrefix(list *protocol.CompletionList, prefix string) {
	for i := range KeywordItems {
		if strings.HasPrefix(KeywordItems[i].Label, prefix) {
			list.Items = append(list.Items, KeywordItems[i])
		}
	}
}
