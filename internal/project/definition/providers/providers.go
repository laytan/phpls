package providers

import (
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/typer"
	"github.com/samber/do"
)

var (
	Index  = func() index.Index { return do.MustInvoke[index.Index](nil) }
	Typer  = func() typer.Typer { return do.MustInvoke[typer.Typer](nil) }
	Wrkspc = func() wrkspc.Wrkspc { return do.MustInvoke[wrkspc.Wrkspc](nil) }
)
