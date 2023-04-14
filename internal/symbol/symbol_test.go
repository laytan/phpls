package symbol_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"appliedgo.net/what"
	"github.com/laytan/phpls/internal/config"
	"github.com/laytan/phpls/internal/index"
	"github.com/laytan/phpls/internal/project"
	"github.com/laytan/phpls/internal/symbol"
	"github.com/laytan/phpls/internal/wrkspc"
	"github.com/laytan/phpls/pkg/pathutils"
	"github.com/laytan/phpls/pkg/phprivacy"
	"github.com/laytan/phpls/pkg/phpversion"
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(
		m,
		// The cache size logger.
		goleak.IgnoreTopFunction("github.com/laytan/phpls/internal/wrkspc.New.func1"),
	)
}

func TestClass(t *testing.T) {
	t.Parallel()

	err := setup(
		filepath.Join(pathutils.Root(), "internal", "symbol", "testdata"),
		phpversion.EightOne(),
	)
	require.NoError(t, err)

	root, err := wrkspc.Current.
		IROf(filepath.Join(pathutils.Root(), "internal", "symbol", "testdata", "test.php"))
	require.NoError(t, err)

	n := root.Stmts[0].(*ast.StmtExpression).Expr.(*ast.ExprNew).Class.(*ast.Name)

	c, err := symbol.NewClassLikeFromName(root, n)
	require.NoError(t, err)

	iter := c.InheritsIter()

	cls, done, err := iter()
	require.False(t, done)
	require.NoError(t, err)
	require.Equal(t, cls.GetFQN().Name(), "TestTrait")

	methIter := cls.MethodsIter()
	meth, done, err := methIter()
	what.Is(meth)
	require.False(t, done)
	require.NoError(t, err)
	require.Equal(t, meth.Name(), "test")
	require.Equal(t, meth.Privacy(), phprivacy.PrivacyPublic)
	require.True(t, meth.CanBeAccessedFrom(phprivacy.PrivacyPublic))
	require.True(t, meth.CanBeAccessedFrom(phprivacy.PrivacyProtected))
	require.True(t, meth.CanBeAccessedFrom(phprivacy.PrivacyPrivate))
	require.False(t, meth.IsStatic())
	require.False(t, meth.IsFinal())

	// zero := iter()
	// is.Equal(zero.FullyQualifier.Get().Name(), "TestBase")
	// plus := iter()
	// is.Equal(plus.FullyQualifier.Get().Name(), "TestBaseInterface")
	// first := iter()
	// is.Equal(first.FullyQualifier.Get().Name(), "TestInterface")
	// scnd := iter()
	// is.Equal(scnd.FullyQualifier.Get().Name(), "TestInterface2")
}

func setup(root string, phpv *phpversion.PHPVersion) error {
	config.Current = config.Default()
	index.Current = index.New(phpv)
	wrkspc.Current = wrkspc.New(
		phpv,
		root,
		filepath.Join(pathutils.Root(), "third_party", "phpstorm-stubs"),
	)

	p := project.New()
	if err := p.ParseWithoutProgress(); err != nil {
		return fmt.Errorf("[symbol_test.setup]: %w", err)
	}

	return nil
}
