package expr

import "fmt"

type RuntimeError struct {
	Err string
}

func (r *RuntimeError) Error() string {
	//TODO implement me
	return r.Err
}

func newErrorf(format string, args ...interface{}) *Error {
	return &Error{Err: &RuntimeError{Err: fmt.Sprintf(format, args...)}}
}
