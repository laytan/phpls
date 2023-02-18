package ie

func IfElse[T any](check bool, ifV T, elseV T) T {
	if check {
		return ifV
	}

	return elseV
}

func IfFuncElse[T any](check bool, ifV func() T, elseV T) T {
	if check {
		return ifV()
	}

	return elseV
}

func IfElseFunc[T any](check bool, ifV T, elseV func() T) T {
	if check {
		return ifV
	}

	return elseV()
}
