package project

import "github.com/laytan/elephp/pkg/phplint"

type Project struct {
	diagnostics map[string][]*phplint.Issue
}

func New() *Project {
	return &Project{
		diagnostics: make(map[string][]*phplint.Issue),
	}
}
