package project

type Issue interface {
	Message() string
	Code() string
	Line() int
}

type Project struct {
	diagnostics map[string][]Issue
}

func New() *Project {
	return &Project{
		diagnostics: make(map[string][]Issue),
	}
}
