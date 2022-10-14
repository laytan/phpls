package phpdoxer

import (
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
		var nameStr, descStr string
		var typeNode Type

		typeOrName, rest := splitTypeAndRest(value)
		typeOrNameRest, desc, _ := strings.Cut(rest, " ")

		if strings.HasPrefix(typeOrName, "$") {
			nameStr = typeOrName
			descStr = rest
		} else if strings.HasPrefix(typeOrNameRest, "$") {
			nameStr = typeOrNameRest
			typeNode, _ = ParseType(typeOrName)
			descStr = desc
		} else {
			descStr = value
		}

		return &NodeParam{
			Type:        typeNode,
			Name:        nameStr,
			Description: descStr,
		}, nil

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
