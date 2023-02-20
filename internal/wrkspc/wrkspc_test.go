package wrkspc_test

import (
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/matryer/is"
	"github.com/samber/do"
)

func BenchmarkWalk(b *testing.B) {
	is := is.New(b)

	do.OverrideValue(nil, config.Default())

	// NOTE: manually change this to some bigger projects when benching.
	do.OverrideValue(nil, wrkspc.New(phpversion.EightOne(), filepath.Join(
		pathutils.Root(),
		"third_party",
		"phpstorm-stubs",
	)))

	for i := 0; i < b.N; i++ {
		filesChan := make(chan *wrkspc.ParsedFile)
		go func() {
			for range filesChan {
			}
		}()

		totalDone := make(chan bool, 1)
		total := &atomic.Uint64{}
		err := wrkspc.FromContainer().Index(filesChan, total, totalDone)
		is.NoErr(err)
	}
}