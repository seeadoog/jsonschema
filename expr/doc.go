package expr

import (
	"fmt"
	"reflect"
	"strings"
)

func GenDocOf(prefix string, v any) string {
	return showDocOf(prefix, v)
}

func showDocOf(prefix, vv any) string {
	v := reflect.ValueOf(vv)
	t := v.Type()

	funs := []string{}
	for i := 0; i < t.NumMethod(); i++ {
		f := v.Method(i)
		ft := f.Type()
		args := []string{}
		for j := 0; j < ft.NumIn(); j++ {
			argi := ft.In(j)
			if argi == contextType && j == 0 {
				continue
			}
			args = append(args, fmt.Sprintf("%v", argi.String()))
		}
		funs = append(funs, fmt.Sprintf("%s%s(%s)", prefix, t.Method(i).Name, strings.Join(args, ",")))
	}
	return strings.Join(funs, "\n")
}
