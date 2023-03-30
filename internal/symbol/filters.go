package symbol

import (
	"github.com/laytan/elephp/pkg/phprivacy"
)

type FilterFunc[T any] func(T) bool

type Named interface {
	Name() string
}

type named interface {
	Named
}

func FilterName[T Named](name string) FilterFunc[T] {
	return func(v T) bool {
		return v.Name() == name
	}
}

type NamedModified interface {
	Named
	Modified
}

func FilterOverwrittenBy[T NamedModified](m T, addDollar ...bool) FilterFunc[T] {
	name := m.Name()
	if len(addDollar) > 0 && addDollar[0] {
		name = "$" + name
	}

	filters := []FilterFunc[T]{
		FilterName[T](name),
		FilterNotPrivacy[T](phprivacy.PrivacyPrivate),
	}

	if !m.IsStatic() {
		filters = append(filters, FilterNotStatic[T]())
	}

	if m.IsStatic() {
		filters = append(filters, FilterStatic[T]())
	}

	return func(m T) bool {
		for _, filter := range filters {
			if !filter(m) {
				return false
			}
		}

		return true
	}
}
