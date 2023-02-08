package phpdoxer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/laytan/elephp/pkg/phpversion"
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

		typeNode, _ := ParseType(typeStr)

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
		if len(split) == 0 {
			return nil, fmt.Errorf(
				"PHPDoc error: found @param %s with 0 arguments, must be at least 1",
				value,
			)
		}

		var name string
		var typeNode Type

		if len(split) == 1 {
			name = split[0]
		}

		if len(split) > 1 {
			name = split[1]

			typeNode, _ = ParseType(split[0])
		}

		return &NodeParam{
			Type: typeNode,
			Name: name,
		}, nil

	case "inheritdoc", "inheritDoc":
		return &NodeInheritDoc{}, nil

	case "since":
		version, description := splitTypeAndRest(value)
		if phpv, ok := phpversion.FromString(version); ok {
			return &NodeSince{
				Version:     phpv,
				Description: description,
			}, nil
		}

		return &NodeUnknown{
			At:    at,
			Value: value,
		}, nil

	case "removed":
		version, description := splitTypeAndRest(value)
		if phpv, ok := phpversion.FromString(version); ok {
			return &NodeRemoved{
				Version:     phpv,
				Description: description,
			}, nil
		}

		return &NodeUnknown{
			At:    at,
			Value: value,
		}, nil

	case "throws":
		typeStr, description := splitTypeAndRest(value)

		typeNode, _ := ParseType(typeStr)

		return &NodeThrows{
			Type:        typeNode,
			Description: description,
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
