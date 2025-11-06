package expr

type Iterator interface {
	Next() (interface{}, bool)
}
