package phpdoxer

import (
	"regexp"
	"strings"

	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/strutil"
)

var (
	groupRgx      = regexp.MustCompile(`@(\w+)\s*([^@]*)`)
	emptyLineRgx  = regexp.MustCompile(`[^*/\s]`)
	whitespaceRgx = regexp.MustCompile(`\n(\s+)`)
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

type Doc struct {
	Top         string
	Indentation string
	Nodes       []Node
}

// Parse the doc string, keeping the leading documentation.
func ParseFullDoc(doc string) (*Doc, error) {
	top, _, _ := strings.Cut(doc, "@")
	nodes, err := ParseDoc(doc)
	if err != nil {
		return nil, err
	}

	indentation := " "
	wsMatches := whitespaceRgx.FindStringSubmatch(doc)
	if len(wsMatches) > 1 {
		indentation = wsMatches[1]
	}

	return &Doc{
		Top:         cleanGroupValue(top),
		Nodes:       nodes,
		Indentation: indentation,
	}, nil
}

// Turns the doc back into a valid PHPDoc string.
//
// NOTE: this is not intended to closely represent the input, but to produce
// NOTE: an output that has all the information the source had,
// NOTE: formatting might be slightly different.
//
// It does try to roughly return the doc with the same indentation level as the
// source.
func (d *Doc) String() string {
	if d.Top == "" && len(d.Nodes) == 0 {
		return ""
	}

	ret := "/**\n"

	for _, line := range strutil.Lines(d.Top) {
		if line != "" {
			ret += d.Indentation + "* " + line + "\n"
		}
	}

	if d.Top != "" && len(d.Nodes) > 0 {
		ret += d.Indentation + "*" + "\n"
	}

	for _, node := range d.Nodes {
		for _, line := range strutil.Lines(node.String()) {
			ret += d.Indentation + "* " + line + "\n"
		}
	}

	ret += d.Indentation + "*/"
	return ret
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
		typeStr, description := splitTypeAndRest(value)
		typeNode, _ := ParseType(typeStr)

		result = &NodeVar{
			Type:        typeNode,
			Description: description,
		}
		return result, nil

	case "param":
		var nameStr, descStr string
		var typeNode Type

		typeOrName, rest := splitTypeAndRest(value)
		typeOrNameRest, desc, _ := strings.Cut(rest, " ")

		//nolint:gocritic // Does not make sense to change to switch.
		if isStrVariable(typeOrName) {
			nameStr = typeOrName
			descStr = rest
		} else if isStrVariable(typeOrNameRest) {
			nameStr = typeOrNameRest
			typeNode, _ = ParseType(typeOrName)
			descStr = desc
		} else {
			descStr = value
		}

		result = &NodeParam{
			Type:        typeNode,
			Name:        nameStr,
			Description: descStr,
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
	lines := strutil.Lines(value)
	outLines := make([]string, 0, len(lines))

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
	if len(split) == 0 {
		return value, ""
	}

	rest = strings.TrimSpace(strings.TrimPrefix(value, split[0]))
	return split[0], rest
}

func isStrVariable(value string) bool {
	return strings.HasPrefix(value, "$") ||
		strings.HasPrefix(value, "...") ||
		strings.HasPrefix(value, "&")
}
