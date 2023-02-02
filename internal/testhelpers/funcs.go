package testhelpers

/********************************************/
/************* Helper Functions *************/
/********************************************/

func PtrTo[T any](val T) *T {
	return &val
}
