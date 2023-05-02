package utils_test

import (
	"testing"

	"github.com/git-ecosystem/git-bundle-server/internal/utils"
	"github.com/stretchr/testify/assert"
)

var segmentKeysTests = []struct {
	title string
	mapA  map[string]interface{}
	mapB  map[string]interface{}

	expectedIntersect []string
	expectedDiffAB    []string
	expectedDiffBA    []string
}{
	{
		"all empty",
		map[string]interface{}{},
		map[string]interface{}{},

		[]string{},
		[]string{},
		[]string{},
	},
	{
		"no overlap",
		map[string]interface{}{
			"A": nil,
			"B": nil,
		},
		map[string]interface{}{
			"C": nil,
			"D": nil,
		},

		[]string{},
		[]string{"A", "B"},
		[]string{"C", "D"},
	},
	{
		"all overlap",
		map[string]interface{}{
			"A": nil,
			"B": nil,
		},
		map[string]interface{}{
			"B": nil,
			"A": nil,
		},

		[]string{"A", "B"},
		[]string{},
		[]string{},
	},
	{
		"A superset of B",
		map[string]interface{}{
			"A": nil,
			"B": nil,
			"C": nil,
			"D": nil,
		},
		map[string]interface{}{
			"B": nil,
			"D": nil,
		},

		[]string{"B", "D"},
		[]string{"A", "C"},
		[]string{},
	},
	{
		"B superset of A",
		map[string]interface{}{
			"A": nil,
			"C": nil,
		},
		map[string]interface{}{
			"A": nil,
			"B": nil,
			"C": nil,
			"D": nil,
		},

		[]string{"A", "C"},
		[]string{},
		[]string{"B", "D"},
	},
	{
		"no empty result sets",
		map[string]interface{}{
			"A": nil,
			"C": nil,
			"E": nil,
		},
		map[string]interface{}{
			"B": nil,
			"C": nil,
			"D": nil,
		},

		[]string{"C"},
		[]string{"A", "E"},
		[]string{"B", "D"},
	},
}

func TestSegmentKeys(t *testing.T) {
	for _, tt := range segmentKeysTests {
		t.Run(tt.title, func(t *testing.T) {
			intersect, diffAB, diffBA := utils.SegmentKeys(tt.mapA, tt.mapB)
			assert.ElementsMatch(t, tt.expectedIntersect, intersect)
			assert.ElementsMatch(t, tt.expectedDiffAB, diffAB)
			assert.ElementsMatch(t, tt.expectedDiffBA, diffBA)
		})
	}
}
