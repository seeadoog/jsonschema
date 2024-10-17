package jsonschema

import (
	"fmt"
	"strings"
)

var valueFuncs = map[string]Func{
	"append": funcAppend,
	"add":    funcAdd,
}

func SetFunc(name string, fun Func) {
	valueFuncs[name] = fun
}

type Context = any

type Value interface {
	Get(ctx Context) interface{}
}

type Const struct {
	Val interface{}
}

func (c *Const) String() string {
	return StringOf(c.Val)
}

func (c Const) Get(ctx Context) interface{} {
	return c.Val
}

type cloneValue struct {
	data map[string]Value
}

func (c *cloneValue) Get(ctx Context) interface{} {
	dst := make(map[string]any, len(c.data))
	for k, v := range c.data {
		dst[k] = v.Get(ctx)
	}
	return dst
}

type sliceValue struct {
	data []Value
}

func (c *sliceValue) Get(ctx Context) interface{} {
	dst := make([]any, len(c.data))
	for k, v := range c.data {
		dst[k] = v.Get(ctx)
	}
	return dst
}

type Var struct {
	Key *JsonPathCompiled
}

func (v Var) Get(ctx Context) interface{} {
	val, err := v.Key.Get(ctx)
	if err != nil {
		return nil
	}
	return val
}

type VarFunc struct {
	funName string
	fn      Func
	args    []Value
}

func (v VarFunc) Get(ctx Context) interface{} {
	return v.fn(ctx, v.args...)

}

type Func func(ctx Context, args ...Value) interface{}

func parseFuncValue(name string, args []interface{}) (Value, error) {
	argsv := make([]Value, len(args))
	for idx, arg := range args {
		argv, err := parseValue(arg)
		if err != nil {
			return nil, err
		}
		argsv[idx] = argv
	}
	f := valueFuncs[name]
	if f == nil {
		return nil, fmt.Errorf("invalid function '%s'", name)
	}
	return &VarFunc{
		funName: name,
		fn:      f,
		args:    argsv,
	}, nil
}

func parseValue(i interface{}) (Value, error) {
	switch i.(type) {
	case map[string]interface{}:
		m := i.(map[string]interface{})
		funName := StringOf(m["func"])
		if valueFuncs[funName] != nil {
			args, ok := m["args"].([]interface{})
			if !ok {
				return &Const{
					Val: i,
				}, nil
			}
			return parseFuncValue(funName, args)
		}

		cv := &cloneValue{
			data: make(map[string]Value),
		}
		for key, val := range m {
			ov, err := parseValue(val)
			if err != nil {
				return nil, err
			}
			cv.data[key] = ov
		}
		return cv, nil

	case string:
		str := i.(string)

		if strings.HasSuffix(str, "()") {
			return parseFuncValue(str[:len(str)-2], nil)
		}
		//if len(str) > 3 && str[0] == '$' && str[1] == '{' && str[len(str)-1] == '}' {
		//	jp, err := parseJpathCompiled(str[2 : len(str)-1])
		//	if err != nil {
		//		return nil, err
		//	}
		//	return &Var{Key: jp}, nil
		//}
		if strings.Contains(str, "${") && strings.Contains(str, "}") {
			return parseComboValue(str)
		}
		return &Const{Val: i}, nil
	case []interface{}:
		vv := i.([]interface{})
		if len(vv) > 0 {
			str := StringOf(vv[0])
			if strings.HasSuffix(str, "()") {
				funcName := str[:len(str)-2]
				args := vv[1:]
				return parseFuncValue(funcName, args)
			}
		}

		sv := &sliceValue{}
		for _, v := range vv {
			svv, err := parseValue(v)
			if err != nil {
				return nil, err
			}
			sv.data = append(sv.data, svv)

		}

		return sv, nil
	default:
		return &Const{Val: i}, nil
	}
}

type comboValue struct {
	values []Value
}

func (v *comboValue) Get(ctx Context) interface{} {

	sb := strings.Builder{}
	for _, val := range v.values {
		sb.WriteString(StringOf(val.Get(ctx)))
	}
	return sb.String()
}

func parseComboValue(s string) (Value, error) {
	token := make([]byte, 0)
	const (
		statusCommon  = 0
		statusVar     = 1
		statusVarScan = 2
	)
	vs := &comboValue{}
	status := statusCommon
	for i := 0; i < len(s); i++ {
		c := s[i]

		switch status {
		case statusCommon:
			token = append(token, c)
			switch c {
			case '$':
				status = statusVar
			}
		case statusVar:
			token = append(token, c)
			if c == '{' {
				if len(token) > 2 {
					vs.values = append(vs.values, &Const{
						Val: string(token[:len(token)-2]),
					})
					token = token[len(token)-2:]
				}
				status = statusVarScan
			} else {
				status = statusCommon
			}

		case statusVarScan:
			token = append(token, c)
			if c == '}' {
				name := string(token[2 : len(token)-1])
				if strings.HasSuffix(name, "()") {
					v, err := parseFuncValue(name[:len(name)-2], nil)
					if err != nil {
						return nil, err
					}
					vs.values = append(vs.values, v)
				} else {
					jp, err := parseJpathCompiled(name)
					if err != nil {
						return nil, err
					}
					vs.values = append(vs.values, &Var{
						Key: jp,
					})
				}

				token = token[:0]
				status = statusCommon
			}

		}
	}
	if len(token) > 0 {
		vs.values = append(vs.values, &Const{
			Val: string(token),
		})
	}

	if len(vs.values) == 1 {
		return vs.values[0], nil
	}
	return vs, nil
}
