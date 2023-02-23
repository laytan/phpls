package throws_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"appliedgo.net/what"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/internal/throws"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/fqn"
	"github.com/laytan/elephp/pkg/functional"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/matryer/is"
	"github.com/samber/do"
	"go.uber.org/goleak"
	"golang.org/x/exp/slices"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(
		m,
		// The cache size logger.
		goleak.IgnoreTopFunction("github.com/laytan/elephp/internal/wrkspc.New.func1"),
	)
}

func TestThrows(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	err := setup(
		filepath.Join(pathutils.Root(), "internal", "throws", "testdata"),
		phpversion.EightOne(),
	)
	is.NoErr(err)

	i := index.FromContainer()

	t.Run("basic one throw using new", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		funcCall, ok := i.Find(fqn.New("\\Throws\\TestData\\test_throws"))
		is.True(ok)

		throws := throws.NewResolverFromIndex(funcCall).Throws()
		is.Equal(len(throws), 1)
		is.Equal(throws[0].String(), "\\Exception")
	})

	t.Run("throw in called function and in current function", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		funcCall, ok := i.Find(fqn.New("\\Throws\\TestData\\test_throws_2"))
		is.True(ok)

		throws := throws.NewResolverFromIndex(funcCall).Throws()
		is.Equal(len(throws), 2)
		ts := functional.Map(throws, func(t *fqn.FQN) string { return t.String() })
		is.True(slices.Contains(ts, "\\Exception"))
		is.True(slices.Contains(ts, "\\Throwable"))
	})

	t.Run("basic catch of same exception class", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		funcCall, ok := i.Find(fqn.New("\\Throws\\TestData\\test_throws_3"))
		is.True(ok)

		throws := throws.NewResolverFromIndex(funcCall).Throws()
		is.Equal(len(throws), 0)
	})

	t.Run("catch of parent class of exception", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		funcCall, ok := i.Find(fqn.New("\\Throws\\TestData\\test_throws_4"))
		is.True(ok)

		throws := throws.NewResolverFromIndex(funcCall).Throws()
		is.Equal(len(throws), 0)
	})

	t.Run("try catch outside of throw scope", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		funcCall, ok := i.Find(fqn.New("\\Throws\\TestData\\test_throws_5"))
		is.True(ok)

		throws := throws.NewResolverFromIndex(funcCall).Throws()
		is.Equal(len(throws), 1)
		is.Equal(throws[0].String(), "\\Exception")
	})

	t.Run("@throws tag", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		funcCall, ok := i.Find(fqn.New("\\Throws\\TestData\\test_throws_6"))
		is.True(ok)

		throws := throws.NewResolverFromIndex(funcCall).Throws()
		is.Equal(len(throws), 1)
		is.Equal(throws[0].String(), "\\Throwable")
	})

	t.Run("@throws tag deep", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		funcCall, ok := i.Find(fqn.New("\\Throws\\TestData\\test_throws_7"))
		is.True(ok)

		throws := throws.NewResolverFromIndex(funcCall).Throws()
		is.Equal(len(throws), 1)
		is.Equal(throws[0].String(), "\\Throwable")
	})

	t.Run("duplicate throw statements", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		funcCall, ok := i.Find(fqn.New("\\Throws\\TestData\\test_throws_8"))
		is.True(ok)

		throws := throws.NewResolverFromIndex(funcCall).Throws()
		is.Equal(len(throws), 1)
		is.Equal(throws[0].String(), "\\InvalidArgumentException")
	})

	t.Run("try 2 statements catch with 1", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		funcCall, ok := i.Find(fqn.New("\\Throws\\TestData\\test_throws_9"))
		is.True(ok)

		throws := throws.NewResolverFromIndex(funcCall).Throws()
		what.Is(throws)
		is.Equal(len(throws), 0)
	})
}

func setup(root string, phpv *phpversion.PHPVersion) error {
	i := index.New(phpv)
	do.OverrideValue(nil, config.Default())
	do.OverrideValue(nil, i)
	do.OverrideValue(
		nil,
		wrkspc.New(phpv, root, filepath.Join(pathutils.Root(), "third_party", "phpstorm-stubs")),
	)

	p := project.New()
	if err := p.ParseWithoutProgress(); err != nil {
		return fmt.Errorf("[throws_test.setup]: %w", err)
	}

	return nil
}
