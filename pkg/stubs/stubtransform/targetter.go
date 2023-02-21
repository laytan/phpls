package stubtransform

import (
	"bytes"
	"fmt"

	"github.com/VKCOM/php-parser/pkg/ast"
	"github.com/laytan/elephp/pkg/functional"
)

// targetter keeps track of the use statements and namespace, ultimately checking
// if a name is the target.
type targetter struct {
	target  [][]byte
	targets [][][]byte
}

func newTargetter(target [][]byte) *targetter {
	return &targetter{
		target:  target,
		targets: [][][]byte{target},
	}
}

func (t *targetter) EnterNamespace(n *ast.StmtNamespace) (exit func()) {
	// If it is a block namespace, like `namespace Test { ... }`, we need to
	// reset the targets.
	var prevTargets [][][]byte
	if n.OpenCurlyBracketTkn != nil {
		prevTargets = t.targets
		t.targets = [][][]byte{t.target}
	}

	return func() {
		// Reset the targets to what they were before this namespace block.
		if prevTargets != nil {
			t.targets = prevTargets
		}
	}
}

func (t *targetter) EnterUse(n *ast.StmtUse) {
	name, _ := n.Use.(*ast.Name)
	if len(name.Parts) != len(t.target) {
		return
	}

	for i, targetPart := range t.target {
		part, ok := name.Parts[i].(*ast.NamePart)
		if !ok {
			return
		}

		if !bytes.Equal(part.Value, targetPart) {
			return
		}
	}

	// We only get here if the use statement is the LanguageLevelTypeAware attribute.

	// No alias, means use the last part of the target name.
	if n.Alias == nil {
		tn := t.target[len(t.target)-1]
		t.targets = append(t.targets, [][]byte{tn})
		return
	}

	// Use the alias name.
	t.targets = append(t.targets, [][]byte{n.Alias.(*ast.Identifier).Value})
}

// Match checks if the given name is a target.
func (t *targetter) Match(n [][]byte) bool {
Targets:
	for _, targetName := range t.targets {
		if len(n) != len(targetName) {
			continue
		}

		for i, targetPart := range targetName {
			if !bytes.Equal(n[i], targetPart) {
				continue Targets
			}
		}

		return true
	}

	return false
}

// MatchName checks if the name of the given node matches any target.
// Only to be called with a *ast.Name or *ast.FullyQualifiedName.
func (t *targetter) MatchName(n ast.Vertex) bool {
	var name [][]byte
	switch tn := n.(type) {
	case *ast.Name:
		name = functional.Map(tn.Parts, func(part ast.Vertex) []byte {
			return part.(*ast.NamePart).Value
		})
	case *ast.NameFullyQualified:
		name = functional.Map(tn.Parts, func(part ast.Vertex) []byte {
			return part.(*ast.NamePart).Value
		})
	default:
		panic(fmt.Sprintf("can't get name of attribute %T", n))
	}

	return t.Match(name)
}
