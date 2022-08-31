package server

type progressKind string

const (
	progressKindBegin  progressKind = "begin"
	progressKindReport progressKind = "report"
	progressKindEnd    progressKind = "end"
)

type progress struct {
	Kind       progressKind `json:"kind,omitempty"`
	Title      string       `json:"title,omitempty"`
	Percentage int          `json:"percentage,omitempty"`
	Message    string       `json:"message,omitempty"`
}
