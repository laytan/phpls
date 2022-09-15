package project

import (
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/samber/do"
)

var (
	Wrkspc = func() wrkspc.Wrkspc { return do.MustInvoke[wrkspc.Wrkspc](nil) }
	Index  = func() index.Index { return do.MustInvoke[index.Index](nil) }
)

type Project struct{}

func New() *Project {
	return &Project{}
}
