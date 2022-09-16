package phpdoxer

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	groupRgx     = regexp.MustCompile(`@(\w+)\s*([^@]*)`)
	emptyLineRgx = regexp.MustCompile(`[^*/\s]`)
)

func ParseDoc(doc string) ([]Node, error) {
	matches := groupRgx.FindAllStringSubmatch(doc, -1)

	parsed := make([]Node, 0, len(matches))
	for _, match := range matches {
		res, err := parseGroup(match[1], match[2])
		if err != nil {
			return nil, err
		}

		parsed = append(parsed, res)
	}

	return parsed, nil
}

func parseGroup(at string, value string) (Node, error) {
	value = cleanGroupValue(value)

	switch at {
	case "return":
		typeStr, description := splitTypeAndRest(value)

		typeNode, err := ParseType(typeStr)
		if err != nil {
			return nil, err
		}

		return &NodeReturn{
			Type:        typeNode,
			Description: description,
		}, nil

	case "var":
		typeStr, _ := splitTypeAndRest(value)

		typeNode, err := ParseType(typeStr)
		if err != nil {
			return nil, err
		}

		return &NodeVar{
			Type: typeNode,
		}, nil

	case "param":
		split := strings.Fields(value)
		if len(split) < 2 {
			return nil, fmt.Errorf(
				"PHPDoc error: found @param %s with %d arguments, must be at least 2 (type and name)",
				value,
				len(split),
			)
		}

		typeStr, name := split[0], split[1]

		typeNode, err := ParseType(typeStr)
		if err != nil {
			return nil, err
		}

		return &NodeParam{
			Type: typeNode,
			Name: name,
		}, nil

	default:
		return &NodeUnknown{
			At:    at,
			Value: value,
		}, nil
	}
}

func cleanGroupValue(value string) string {
	lines := strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
	outLines := []string{}

	// Removes any leading or ending /'s, *'s or whitespace.
	// Removes any empty lines.
	for _, line := range lines {
		if !emptyLineRgx.MatchString(line) {
			continue
		}

		line = strings.TrimSpace(line)
		line = strings.Trim(line, "/*")
		line = strings.TrimSpace(line)

		outLines = append(outLines, line)
	}

	return strings.Join(outLines, "\n")
}

func splitTypeAndRest(value string) (docType string, rest string) {
	split := strings.Fields(value)
	rest = strings.TrimSpace(strings.TrimPrefix(value, split[0]))
	return split[0], rest
}
