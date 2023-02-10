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

type group struct {
	full     string
	at       string
	value    string
	startPos int
	endPos   int
}

func ParseDoc(doc string) ([]Node, error) {
	groups := []*group{}
	iMatches := groupRgx.FindAllStringSubmatchIndex(doc, -1)
	for _, iMatch := range iMatches {
		groups = append(groups, &group{
			full:     doc[iMatch[0]:iMatch[1]],
			at:       doc[iMatch[2]:iMatch[3]],
			value:    doc[iMatch[4]:iMatch[5]],
			startPos: iMatch[0],
			endPos:   iMatch[1],
		})
	}

	parsed := make([]Node, 0, len(groups))
	for _, g := range groups {
		res, err := parseGroup(g)
		if err != nil {
			return nil, err
		}

		parsed = append(parsed, res)
	}

	return parsed, nil
}

func parseGroup(g *group) (Node, error) {
	var result Node

	// Add the range to the result node before returning.
	defer func() {
		if result != nil {
			result.setRange(g.startPos, g.endPos)
		}
	}()

	value := cleanGroupValue(g.value)

	switch g.at {
	case "return":
		typeStr, description := splitTypeAndRest(value)

		typeNode, _ := ParseType(typeStr)

		result = &NodeReturn{
			Type:        typeNode,
			Description: description,
		}
		return result, nil

	case "var":
		typeStr, _ := splitTypeAndRest(value)

		typeNode, err := ParseType(typeStr)
		if err != nil {
			return nil, err
		}

		result = &NodeVar{
			Type: typeNode,
		}
		return result, nil

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

		result = &NodeParam{
			Type: typeNode,
			Name: name,
		}
		return result, nil

	case "inheritdoc", "inheritDoc":
		result = &NodeInheritDoc{}
		return result, nil

	case "since":
		version, description := splitTypeAndRest(value)
		if phpv, ok := phpversion.FromString(version); ok {
			result = &NodeSince{
				Version:     phpv,
				Description: description,
			}
			return result, nil
		}

		result = &NodeUnknown{
			At:    g.at,
			Value: value,
		}
		return result, nil

	case "removed":
		version, description := splitTypeAndRest(value)
		if phpv, ok := phpversion.FromString(version); ok {
			result = &NodeRemoved{
				Version:     phpv,
				Description: description,
			}
			return result, nil
		}

		result = &NodeUnknown{
			At:    g.at,
			Value: value,
		}
		return result, nil

	case "throws":
		typeStr, description := splitTypeAndRest(value)

		typeNode, _ := ParseType(typeStr)

		result = &NodeThrows{
			Type:        typeNode,
			Description: description,
		}
		return result, nil

	default:
		result = &NodeUnknown{
			At:    g.at,
			Value: value,
		}
		return result, nil
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
