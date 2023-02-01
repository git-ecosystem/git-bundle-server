package testhelpers

/*********************************************/
/************* Types & Functions *************/
/*********************************************/

type Pair[T any, R any] struct {
	First  T
	Second R
}

func NewPair[T any, R any](first T, second R) Pair[T, R] {
	return Pair[T, R]{
		First:  first,
		Second: second,
	}
}

type BoolArg int

const (
	False BoolArg = iota
	True
	Any
)

func (b BoolArg) ToBoolList() []bool {
	switch b {
	case False:
		return []bool{false}
	case True:
		return []bool{true}
	case Any:
		return []bool{false, true}
	default:
		panic("invalid bool arg value")
	}
}
