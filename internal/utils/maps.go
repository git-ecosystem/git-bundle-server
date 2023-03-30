package utils

// SegmentKeys takes two maps with keys of the same type and returns three
// slices (with *no guaranteed order*):
// - the keys common across both maps
// - the keys present in the first set but not the second
// - the keys present in the second set but not the first
func SegmentKeys[T comparable, S any, R any](mapA map[T]S, mapB map[T]R) (intersect, diffAB, diffBA []T) {
	intersect = []T{}
	diffAB = []T{}
	diffBA = []T{}
	for a := range mapA {
		if _, contains := mapB[a]; contains {
			intersect = append(intersect, a)
		} else {
			diffAB = append(diffAB, a)
		}
	}

	for b := range mapB {
		if _, contains := mapA[b]; !contains {
			diffBA = append(diffBA, b)
		}
	}

	return
}
