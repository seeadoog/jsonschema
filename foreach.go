package jsonschema

import "fmt"

type forEach struct {
	//does map[*JsonPathCompiled]Validator
	does sliceMap[*JsonPathCompiled, Validator]
}

func (f *forEach) Validate(ctx *ValidateCtx, pv any) {
	vm, ok := pv.(map[string]any)
	if !ok {
		return
	}
	f.does.Range(func(jp *JsonPathCompiled, validator Validator) bool {
		val, err := jp.Get(pv)
		if err != nil {
			return true
		}

		switch val := val.(type) {
		case []any:
			for i, v := range val {
				vm["__key"] = float64(i)
				vm["__val"] = v
				validator.Validate(ctx, pv)
			}
		case map[string]any:
			for k, v := range val {
				vm["__key"] = k
				vm["__val"] = v
				validator.Validate(ctx, pv)
			}
		}
		delete(vm, "__key")
		delete(vm, "__val")
		return true
	})

}

var newForeach NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {
	mm, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s type should be a map", path)
	}
	fe := &forEach{
		//does: make(map[*JsonPathCompiled]Validator),
	}
	for name, val := range mm {
		vp, err := NewProp(val, path+"."+name)
		if err != nil {
			return nil, fmt.Errorf("%s parse foreach as prop err :%w", path, err)
		}
		jp, err := parseJpathCompiled(name)
		if err != nil {
			return nil, fmt.Errorf("%s.%s parse foreach as jsonpath err :%w", path, name, err)
		}
		//fe.does[jp] = vp
		fe.does.Set(jp, vp)
	}
	return fe, nil
}

type mapOperation struct {
	op  func(m map[string]any, key string, val any)
	key Value
	val Value
}

func (v *mapOperation) Validate(ctx *ValidateCtx, pv any) {
	mm, ok := pv.(map[string]any)
	if !ok {
		return
	}
	var vvv any
	if v.val != nil {
		vvv = v.val.Get(mm)
	}
	kkk := v.key.Get(mm)
	v.op(mm, StringOf(kkk), vvv)
}

func NewMapOpt(op func(m map[string]any, key string, val any)) NewValidatorFunc {
	return func(i interface{}, path string, parent Validator) (Validator, error) {
		mm, ok := i.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%s type should be a map", path)
		}
		kv, err := parseValue(mm["key"])
		if err != nil {
			return nil, fmt.Errorf("%s parse setMap.key err :%w", path, err)
		}
		vv, err := parseValue(mm["val"])
		if err != nil {
			return nil, fmt.Errorf("%s parse setMap.val err :%w", path, err)
		}

		return &mapOperation{op: op, key: kv, val: vv}, nil
	}
}
