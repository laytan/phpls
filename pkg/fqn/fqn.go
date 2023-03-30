package fqn

import (
	"log"
	"strings"
)

const PartSeperator = `\`

type FQN struct {
	// Examples: \DateTime, \Test\DateTime.
	value string
}

func New(value string) *FQN {
	if value[0:1] != PartSeperator {
		value = PartSeperator + value
		log.Printf(
			"[TODO]: Find where no suffix of \\ is added: Trying to create FQN without a fully qualified input: %q.",
			value,
		) // TODO
	}

	r := &FQN{value: value}
	return r
}

func (f *FQN) String() string {
	return f.value
}

func (f *FQN) Namespace() string {
	parts := f.Parts()
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts[:len(parts)-1], PartSeperator)
}

func (f *FQN) Name() string {
	parts := f.Parts()
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}

func (f *FQN) WithoutHead() *FQN {
	return New("\\" + f.Namespace())
}

func (f *FQN) Parts() []string {
	if f.value == PartSeperator {
		return []string{}
	}

	parts := strings.Split(f.value, PartSeperator)
	if len(parts) == 0 {
		return parts
	}

	return parts[1:]
}
