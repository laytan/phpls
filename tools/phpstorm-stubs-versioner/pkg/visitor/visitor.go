package visitor

type Logger interface {
	Printf(format string, args ...any)
}
