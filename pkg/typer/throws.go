package typer

import (
	"fmt"

	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/pkg/phpdoxer"
	"github.com/laytan/elephp/pkg/queue"
)

func (t *typer) Throws(funcOrMeth ir.Node) (throws []phpdoxer.Type) {
	kind := ir.GetNodeKind(funcOrMeth)
	if kind != ir.KindClassMethodStmt && kind != ir.KindFunctionStmt {
		panic(fmt.Errorf("Type: %T: %w", funcOrMeth, ErrUnexpectedNodeType))
	}

	cmntQueue := queue.New[phpdoxer.Node]()
	addFuncComments(cmntQueue, funcOrMeth)

	for cmnt := cmntQueue.Dequeue(); cmnt != nil; cmnt = cmntQueue.Dequeue() {
		if cmnt.Kind() == phpdoxer.KindThrows {
			throws = append(throws, cmnt.(*phpdoxer.NodeThrows).Type)
		}
	}

	return throws
}
