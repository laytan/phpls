package phplint_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/laytan/elephp/pkg/pathutils"
	"github.com/laytan/elephp/pkg/phplint"
	"github.com/matryer/is"
)

var _, isCI = os.LookupEnv("CI")

func skipCI(t *testing.T) {
	t.Helper()

	if isCI {
		t.Skip(
			"Requires PHP to be installed, which I don't want to do on CI for this one test suite",
		)
	}
}

func TestPhpLint(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		skipCI(t)

		out, err := phplint.LintString([]byte(""))
		is.NoErr(err)
		if len(out) != 0 {
			t.Errorf(
				"Empty string passed to lint should return 0 issues but returned %d, got: %v",
				len(out),
				out,
			)
		}
	})

	t.Run("one line incomplete", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		skipCI(t)

		out, err := phplint.LintString([]byte("<?php echo ?>"))
		is.NoErr(err)
		if len(out) == 0 || out[0].Line != 1 {
			t.Errorf(
				"Error linting string '<?php echo ?>', should return an issue on line 1, got: %v",
				out,
			)
		}
	})

	t.Run("multi line weird whiles", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		skipCI(t)

		out, err := phplint.LintString([]byte(`
<?php
    do
     }   echo 'dowhile test';
    while (false);
?>`,
		))
		is.NoErr(err)
		if len(out) == 0 {
			t.Errorf(
				"Error linting string, should return an issue, got: %v",
				out,
			)
		}
	})
}

func TestPhpLintFile(t *testing.T) {
	t.Parallel()

	if isCI {
		t.Skip(
			"Requires PHP to be installed, which I don't want to do on CI for this one test suite",
		)
	}

	t.Run("syntax_errors.php", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		skipCI(t)

		out, err := phplint.LintFile(
			filepath.Join(
				pathutils.Root(),
				"test",
				"testdata",
				"syntaxerrors",
				"syntax_errors.php",
			),
		)
		is.NoErr(err)
		if len(out) == 0 {
			t.Errorf(
				"Error linting string, should return an issue, got: %v",
				out,
			)
		}
	})

	t.Run("bad_whiles.php", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		skipCI(t)

		out, err := phplint.LintFile(
			filepath.Join(pathutils.Root(), "test", "testdata", "syntaxerrors", "bad_whiles.php"),
		)
		is.NoErr(err)
		if len(out) == 0 {
			t.Errorf(
				"Error linting string, should return an issue, got: %v",
				out,
			)
		}
	})
}
