package project

import (
	"github.com/laytan/elephp/internal/index"
	"github.com/laytan/elephp/internal/wrkspc"
	"github.com/laytan/elephp/pkg/phpversion"
	"github.com/laytan/elephp/pkg/typer"
)

type Project struct {
	index index.Index
	wrksp wrkspc.Wrkspc
	typer typer.Typer
}

func New(root string, version *phpversion.PHPVersion, fileExtensions []string) *Project {
	return &Project{
		index: index.New(version),
		wrksp: wrkspc.New(version, root, fileExtensions),
		typer: typer.New(),
	}
}
