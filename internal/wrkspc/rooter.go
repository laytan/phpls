package wrkspc

import (
    "log"

	"github.com/VKCOM/noverify/src/ir"
)

type Rooter struct {
	path string
	root *ir.Root
}

func NewRooter(path string, root ...*ir.Root) *Rooter {
	if len(root) > 1 {
		log.Panic("[wrkspc.NewRooter]: can only create a rooter with one root")
	}

	if len(root) == 1 {
		return &Rooter{path: path, root: root[0]}
	}

	return &Rooter{path: path}
}

func (r *Rooter) Path() string { return r.path }
func (r *Rooter) Root() *ir.Root {
	if r.root == nil {
		r.root = FromContainer().FIROf(r.path)
	}

	return r.root
}
