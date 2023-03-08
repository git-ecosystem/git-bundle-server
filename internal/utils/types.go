package utils

// KeyValue is very similar to testhelpers.Pair. Consider unifying on a single
// 'utils.Pair' if https://github.com/golang/go/issues/52654 is ever
// implemented. With that, we can alias 'utils.NewPair' in 'testhelpers' with:
//
// func NewPair = utils.NewPair
//
// and avoid the need to either prefix all 'NewPair' calls with 'utils.' (since
// we don't want to dot-import 'utils' either).

type KeyValue[T any, R any] struct {
	Key   T
	Value R
}

func NewKeyValue[T any, R any](key T, value R) KeyValue[T, R] {
	return KeyValue[T, R]{
		Key:   key,
		Value: value,
	}
}
