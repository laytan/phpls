package phpdoxer

import (
	"regexp"
	"strings"
)

var allRegex = regexp.MustCompile(`@(\w+) ?(.*)`)

type PhpDoxer interface {
	All(string) map[string]string
	Var(string) []string
	Return(string) []string
}

type phpdoxer struct{}

func (p *phpdoxer) All(doc string) map[string]string {
	result := make(map[string]string)
	matches := allRegex.FindAllStringSubmatch(doc, -1)
	for _, match := range matches {
		if len(match) < 2 {
			result[match[1]] = ""
			continue
		}

		result[match[1]] = match[2]
	}

	return result
}

func (p *phpdoxer) Var(doc string) []string {
	result := []string{}
	for k, v := range p.All(doc) {
		if k != "var" || len(v) == 0 {
			continue
		}

		parts := strings.Split(v, " ")
		if len(parts) > 0 {
			result = strings.Split(parts[0], "|")
		}
	}

	return convertNulls(result)
}

func (p *phpdoxer) Return(doc string) []string {
	result := []string{}
	for k, v := range p.All(doc) {
		if k != "return" || len(v) == 0 {
			continue
		}

		parts := strings.Split(v, " ")
		if len(parts) > 0 {
			result = strings.Split(parts[0], "|")
		}
	}

	return convertNulls(result)
}

func convertNulls(result []string) []string {
	var hasNull bool
	for i, v := range result {
		if v == "null" {
			if hasNull {
				result = append(result[:i], result[i+1:]...)
				continue
			}

			hasNull = true
		}

		if strings.HasPrefix(v, "?") {
			result[i] = strings.TrimPrefix(result[i], "?")
			if hasNull {
				continue
			}

			result = append(result, "null")
			hasNull = true
		}
	}

	return result
}
