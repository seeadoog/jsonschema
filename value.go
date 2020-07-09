package jsonschema


var valueFuncs = map[string]Func{
	"append": funcAppend,
	"add":    funcAdd,
}

func SetFunc(name string, fun Func) {
	valueFuncs[name] = fun
}

type Context map[string]interface{}
type Value interface {
	Get(ctx Context) interface{}
}

type Const struct {
	Val interface{}
}

func (c Const) Get(ctx Context) interface{} {
	return c.Val
}

type Var struct {
	Key *JsonPathCompiled
}

func (v Var) Get(ctx Context) interface{} {
	val, err := v.Key.Get(map[string]interface{}(ctx))
	if err != nil {
		return nil
	}
	return val
}

type VarFunc struct {
	funName string
	args    []Value
}

func (v VarFunc) Get(ctx Context) interface{} {
	fun := valueFuncs[v.funName]
	if fun == nil {
		return nil
	}
	return fun(ctx, v.args...)

}

type Func func(ctx Context, args ...Value) interface{}

func parseValue(i interface{}) (Value, error) {
	switch i.(type) {
	case map[string]interface{}:
		m := i.(map[string]interface{})
		funName := String(m["func"])
		//from func
		fv := &VarFunc{
			funName: funName,
		}
		args, ok := m["args"].([]interface{})
		if !ok {
			return fv, nil
		}
		argsv := make([]Value, len(args))
		for idx, arg := range args {
			argv, err := parseValue(arg)
			if err != nil {
				return nil, err
			}
			argsv[idx] = argv
		}
		fv.args = argsv
		return fv, nil

	case string:
		str := i.(string)
		if len(str) > 3 && str[0] == '$' && str[1] == '{' && str[len(str)-1] == '}' {
			jp, err := parseJpathCompiled(str[2 : len(str)-1])
			if err != nil {
				return nil, err
			}
			return Var{Key: jp}, nil
		}
		return Const{Val: i}, nil
	default:
		return &Const{Val: i}, nil
	}
}
