package wrkspc_test

import (
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/stretchr/testify/require"
)

func BenchmarkWalk(b *testing.B) {
	config.Current = config.Default()

	stubs := filepath.Join(pathutils.Root(), "third_party", "phpstorm-stubs")
	// NOTE: manually change this to some bigger projects when benching.
	wrkspc.Current = wrkspc.New(phpversion.EightOne(), stubs, stubs)

	for i := 0; i < b.N; i++ {
		filesChan := make(chan *wrkspc.ParsedFile)
		go func() {
			for range filesChan {
			}
		}()

		totalDone := make(chan bool, 1)
		total := &atomic.Uint64{}
		err := wrkspc.Current.Index(filesChan, total, totalDone)
		require.NoError(b, err)
	}
}
