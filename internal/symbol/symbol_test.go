package symbol_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"appliedgo.net/what"
	"github.com/VKCOM/noverify/src/ir"
	"github.com/laytan/elephp/internal/config"
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/project"
	"github.com/laytan/elephp/internal/symbol"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phprivacy"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/typer"
	"github.com/matryer/is"
	"github.com/samber/do"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(
		m,
		// The cache size logger.
		goleak.IgnoreTopFunction("github.com/laytan/elephp/internal/wrkspc.New.func1"),
	)
}

func TestClass(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	err := setup(
		filepath.Join(pathutils.Root(), "internal", "symbol", "testdata"),
		phpversion.EightOne(),
	)
	is.NoErr(err)

	root, err := wrkspc.FromContainer().
		IROf(filepath.Join(pathutils.Root(), "internal", "symbol", "testdata", "test.php"))
	is.NoErr(err)

	n := root.Stmts[0].(*ir.ExpressionStmt).Expr.(*ir.NewExpr).Class.(*ir.Name)

	c, err := symbol.NewClassLikeFromName(root, n)
	is.NoErr(err)

	iter := c.InheritsIter()

	cls, done, err := iter()
	is.Equal(done, false)
	is.NoErr(err)
	is.Equal(cls.GetFQN().Name(), "TestTrait")

	methIter := cls.MethodsIter()
	meth, done, err := methIter()
	what.Is(meth)
	is.Equal(done, false)
	is.NoErr(err)
	is.Equal(meth.Name(), "test")
	is.Equal(meth.Privacy(), phprivacy.PrivacyPublic)
	is.True(meth.CanBeAccessedFrom(phprivacy.PrivacyPublic))
	is.True(meth.CanBeAccessedFrom(phprivacy.PrivacyProtected))
	is.True(meth.CanBeAccessedFrom(phprivacy.PrivacyPrivate))
	is.Equal(meth.IsStatic(), false)
	is.Equal(meth.IsFinal(), false)

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
	i := index.New(phpv)
	do.OverrideValue(nil, config.Default())
	do.OverrideValue(nil, i)
	do.OverrideValue(nil, wrkspc.New(phpv, root))
	do.OverrideValue(nil, typer.New())

	p := project.New()
	if err := p.ParseWithoutProgress(); err != nil {
		return fmt.Errorf("[symbol_test.setup]: %w", err)
	}

	return nil
}
