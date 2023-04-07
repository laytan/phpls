package phpcbf_test

import (
	"os"
	"testing"

	"github.com/laytan/elephp/pkg/phpcs/phpcbf"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestPhpcbf(t *testing.T) {
	t.Parallel()
	if _, ok := os.LookupEnv("CI"); ok {
		t.Skip() // Skip because I don't want to install phpcbf on ci.
	}

	p := phpcbf.NewInstance()
	p.Init()
	defer p.Close()

	in := []byte(`<?php add();`)
	expected := []byte(`<?php add();`)
	formatted, err := p.Format(in)
	require.NoError(t, err)
	require.Equal(t, string(expected), string(formatted))
}
