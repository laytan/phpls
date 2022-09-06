package index

import (
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/traversers"
)

type Kind int

const (
	KindClass Kind = iota
	KindInterface
	KindTrait
	KindFunction
	KindGlobalVariable
)

type Index interface {
	Index(path string, content string) error

	IndexIncoming(files <-chan *wrkspc.ParsedFile, done <-chan bool) error

	Find(FQN string, kind ...Kind) []*traversers.TrieNode
}
