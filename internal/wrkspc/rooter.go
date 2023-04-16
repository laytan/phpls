package wrkspc

import (
	"log"

	"github.com/laytan/php-parser/pkg/ast"
)

// Rooter is a common way of passing around the root and path of a file/symbol.
//
// If root is not available upon creation, you can leave it nil and it will be
// retrieved and cached when asked for.
//
// A valid path is mandatory though.
type Rooter struct {
	path string
	root *ast.Root
}

// NewRooter creates a new rooter struct, you can pass 0 or 1 root.
// If 0 roots, the root will be retrieved and cached when asked for.
// If 1 root, the given root is used when asked for.
// If > 2 roots, panic.
func NewRooter(path string, root ...*ast.Root) *Rooter {
	if len(root) > 1 {
		log.Panic("[wrkspc.NewRooter]: can only create a rooter with one root")
	}

	if len(root) == 1 {
		return &Rooter{path: path, root: root[0]}
	}

	return &Rooter{path: path}
}

func (r *Rooter) Path() string { return r.path }
func (r *Rooter) Root() *ast.Root {
	if r.root == nil {
		r.root = Current.AstF(r.path)
	}

	return r.root
}
