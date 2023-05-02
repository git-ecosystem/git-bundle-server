package utils_test

import (
	"strings"
	"testing"

	"github.com/git-ecosystem/git-bundle-server/internal/utils"
	"github.com/stretchr/testify/assert"
)

// To test generics, implement the general-purpose 'testable' interface in
// a templated, function-specific struct (like 'mapTest') that define how
// the associated generic function is tested.

type testable interface {
	runTest(t *testing.T)
}

type mapTest[T any, S any] struct {
	title string

	// Inputs
	in []T
	fn func(T) S

	// Outputs
	expectedOut []S
}

type twoInts struct {
	int1 int
	int2 int
}

var mapTests = []testable{
	mapTest[string, string]{
		title: "string -> string",

		in: []string{"  A ", "B\t", "\nC  \r\n", "D", "    E\t"},
		fn: strings.TrimSpace,

		expectedOut: []string{"A", "B", "C", "D", "E"},
	},
	mapTest[int, float32]{
		title: "int -> float32",

		in: []int{1, 2, 3, 4, 5},
		fn: func(i int) float32 { return float32(i) },

		expectedOut: []float32{1, 2, 3, 4, 5},
	},
	mapTest[string, struct{ name string }]{
		title: "string -> anonymous struct",

		in: []string{"test", "another test"},
		fn: func(str string) struct{ name string } { return struct{ name string }{name: str} },

		expectedOut: []struct{ name string }{{name: "test"}, {name: "another test"}},
	},
	mapTest[twoInts, int]{
		title: "named struct -> int",

		in: []twoInts{{int1: 12, int2: 34}, {int1: 56, int2: 78}},
		fn: func(s twoInts) int { return s.int1 + s.int2 },

		expectedOut: []int{46, 134},
	},
}

func (tt mapTest[T, S]) runTest(t *testing.T) {
	t.Run(tt.title, func(t *testing.T) {
		out := utils.Map(tt.in, tt.fn)
		assert.Equal(t, tt.expectedOut, out)
	})
}

func TestMap(t *testing.T) {
	for _, tt := range mapTests {
		tt.runTest(t)
	}
}

type reduceTest[T any, S any] struct {
	title string

	// Inputs
	in  []T
	acc S
	fn  func(T, S) S

	// Outputs
	expectedOut S
}

var reduceTests = []testable{
	reduceTest[int, int]{
		title: "int -> int",

		in:  []int{1, 2, 3},
		acc: 0,
		fn:  func(t int, acc int) int { return acc + t },

		expectedOut: 6,
	},
	reduceTest[string, []string]{
		title: "string -> []string",

		in:  []string{"A", "B", "C"},
		acc: []string{"test"},
		fn:  func(t string, acc []string) []string { return append(acc, t) },

		expectedOut: []string{"test", "A", "B", "C"},
	},
	reduceTest[twoInts, int]{
		title: "named struct -> int",

		in:  []twoInts{{int1: 12, int2: 34}, {int1: 56, int2: 78}},
		acc: 0,
		fn:  func(t twoInts, acc int) int { return acc + t.int1 + t.int2 },

		expectedOut: 180,
	},
}

func (tt reduceTest[T, S]) runTest(t *testing.T) {
	t.Run(tt.title, func(t *testing.T) {
		out := utils.Reduce(tt.in, tt.acc, tt.fn)
		assert.Equal(t, tt.expectedOut, out)
	})
}

func TestReduce(t *testing.T) {
	for _, tt := range reduceTests {
		tt.runTest(t)
	}
}
