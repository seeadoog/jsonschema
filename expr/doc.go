package expr

import (
	"fmt"
	"reflect"
	"strings"
)

func GenDocOf(prefix string, v any) string {
	return showDocOf(prefix, v)
}

func docOfFunc(prefix string, ft reflect.Type, vt reflect.Type) (funs []string) {
	args := []string{}
	tms := []string{}

	tmp := map[reflect.Type]bool{}
	for j := 0; j < ft.NumIn(); j++ {
		argi := ft.In(j)
		if argi == contextType && j == 0 {
			continue
		}
		args = append(args, fmt.Sprintf("%v", argi.String()))

		if !tmp[argi] {
			tmp[argi] = true
			tms = append(tms, argi.String())
			tms = append(tms, docOfStruct(prefix, ft.In(j), ft.In(j))...)
		}

	}
	funs = append(funs, fmt.Sprintf("%s%s(%s)", prefix, vt.Name(), strings.Join(args, ",")))
	funs = append(funs, tms...)
	return funs
}

func docOfStruct(prefix string, v reflect.Type, t reflect.Type) (funs []string) {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	if v.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			ft := t.Field(i)
			if !ft.IsExported() {
				continue
			}
			funs = append(funs, fmt.Sprintf("%s%s:%s", prefix, ft.Name, ft.Type.String()))
		}
	}
	return funs
}

func showDocOf(prefix string, vv any) string {
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

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			ft := t.Field(i)
			if !ft.IsExported() {
				continue
			}
			funs = append(funs, fmt.Sprintf("%s%s:%s", prefix, ft.Name, ft.Type.String()))
		}
	}

	if v.Kind() == reflect.Func {
		funs = append(funs, docOfFunc(prefix, t, t)...)
	}
	return strings.Join(funs, "\n")
}
