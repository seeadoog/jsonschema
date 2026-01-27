package expr

import (
	"fmt"
	"reflect"
)

func newCompileFunc(args []Val, argsNum int, cf compileFunc) (newArgs []Val, err error) {
	cargs := make([]any, 0, argsNum)
	for i, val := range args[:argsNum] {
		constVal, ok := val.(*constraint)
		if !ok {
			return nil, fmt.Errorf(" args: %v is not a constraint,but %v", i, reflect.TypeOf(val))
		}
		cargs = append(cargs, constVal.value)
	}
	res, err := cf(cargs...)
	if err != nil {
		return nil, err
	}
	newArgs = make([]Val, len(args))
	copy(newArgs, args)

	newArgs[0] = &constraint{value: res}
	return newArgs, nil
}
