package testhelpers

import "github.com/github/git-bundle-server/internal/utils"

/********************************************/
/************* Helper Functions *************/
/********************************************/

func PtrTo[T any](val T) *T {
	return &val
}

func ConcatLines(lines []string) string {
	return utils.Reduce(lines, "", func(line string, out string) string {
		return out + line + "\n"
	})
}
