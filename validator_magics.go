package jsonschema

import (
	"fmt"
	"strings"
)

type ConstVal struct {
	Val interface{}
}

func (cc ConstVal) Validate(c *ValidateCtx, value interface{}) {

}

type DefaultVal struct {
	Val interface{}
}

func (d DefaultVal) Validate(c *ValidateCtx, value interface{}) {

}

type ReplaceKey string

func (r ReplaceKey) Validate(c *ValidateCtx, value interface{}) {

}

func NewConstVal(i interface{}, path string, parent Validator) (Validator, error) {
	return &ConstVal{
		Val: i,
	}, nil
}

func NewDefaultVal(i interface{}, path string, parent Validator) (Validator, error) {
	return &DefaultVal{i}, nil
}

func NewReplaceKey(i interface{}, path string, parent Validator) (Validator, error) {
	s, ok := i.(string)
	if !ok {
		return nil, fmt.Errorf("value of 'replaceKey' must be string :%v", i)
	}
	return ReplaceKey(s), nil

}

type FormatVal _type

func (f FormatVal) Validate(c *ValidateCtx, value interface{}) {

}

func (f FormatVal) Convert(value interface{}) interface{} {
	switch _type(f) {
	case typeString:
		return StringOf(value)
	case typeBool:
		return BoolOf(value)
	case typeInteger, typeNumber:
		return NumberOf(value)
	case typeLower:
		return strings.ToLower(StringOf(value))
	case typeUpper:
		return strings.ToUpper(StringOf(value))
	}
	return value
}

func NewFormatVal(i interface{}, path string, parent Validator) (Validator, error) {
	str, ok := i.(string)
	if !ok {
		return nil, fmt.Errorf("value of format must be string:%s", str)
	}
	return FormatVal(types[str]), nil
}

/*
{
	"setVal":{
		"key1":1,
		"key2":"val2",
		"key3":"${key1}",
		"key4":{
			"func":"append",
			"args":["${key1}","${key2}",{"func":"add","args":[1,2]}]
		},
	}
}
{
	"if":{
		"op":"eq",
		"l":"",
		"r":""
	}
	"then":{

	},

	"else":{

	},
	"and":[
		{
			"if":{}
		}
	],
	"set":{
		"k1":"",


	}
}

*/

type SetVal struct {
	data          sliceMap[*JsonPathCompiled, Value]
	setWithoutNil bool
}

func (s *SetVal) Validate(c *ValidateCtx, value interface{}) {

	ctx := value
	s.data.Range(func(key *JsonPathCompiled, val Value) bool {
		v := val.Get(ctx)
		if s.setWithoutNil && v == nil {
			return true
		}
		key.Set(ctx, v)
		return true
	})

}

func NewSetVal(i interface{}, path string, parent Validator) (Validator, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s value of setVal must be map[string]interface{} :%v", path, i)
	}

	setVal := SetVal{}
	for key, val := range m {
		v, err := parseValue(val)
		if err != nil {
			return nil, err
		}
		jp, err := parseJpathCompiled(key)
		if err != nil {
			return nil, err
		}
		//setVal[jp] = v
		setVal.data.Set(jp, v)
	}
	return &setVal, nil
}

func NewSetValWithoutNil(i interface{}, path string, parent Validator) (Validator, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s value of setNN must be map[string]interface{} :%v", path, i)
	}

	setVal := SetVal{
		setWithoutNil: true,
	}
	for key, val := range m {
		v, err := parseValue(val)
		if err != nil {
			return nil, err
		}
		jp, err := parseJpathCompiled(key)
		if err != nil {
			return nil, err
		}
		//setVal[jp] = v
		setVal.data.Set(jp, v)
	}
	return &setVal, nil
}

func NewWithSlice(f NewValidatorFunc) NewValidatorFunc {
	return func(i interface{}, path string, parent Validator) (Validator, error) {

		switch arr := i.(type) {
		case []interface{}:
			all := AllOf{}
			for _, o := range arr {
				v, err := f(o, path, parent)
				if err != nil {
					return nil, err
				}
				all = append(all, v)
			}
			return all, nil
		default:
			return f(i, path, parent)
		}
	}
}

type SetExpr struct {
	data sliceMap[Value, Value]
}

func (s *SetExpr) Validate(c *ValidateCtx, value interface{}) {
	m, ok := value.(map[string]interface{})
	if !ok {
		return
	}
	ctx := Context(m)
	s.data.Range(func(key Value, val Value) bool {
		v := val.Get(ctx)

		k := key.Get(ctx)
		m[StringOf(k)] = v
		//key.Set(m, v)
		return true
	})

}

func NewSetExpr(i interface{}, path string, parent Validator) (Validator, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s value of setVal must be map[string]interface{} :%v", path, i)
	}

	setVal := SetExpr{}
	for key, val := range m {
		v, err := parseValue(val)
		if err != nil {
			return nil, err
		}
		jp, err := parseValue(key)
		if err != nil {
			return nil, err
		}
		//setVal[jp] = v
		setVal.data.Set(jp, v)
	}
	return &setVal, nil
}

type setTo struct {
	data sliceMap[*JsonPathCompiled, *sliceMap[Value, Value]]
}

func (s *setTo) Validate(c *ValidateCtx, value interface{}) {
	s.data.Range(func(k *JsonPathCompiled, v *sliceMap[Value, Value]) bool {
		obj, err := k.Get(value)
		if err != nil {
			return true
		}
		om, ok := obj.(map[string]any)
		if !ok {
			return true
		}
		v.Range(func(k Value, v Value) bool {
			key := k.Get(value)
			val := v.Get(value)
			om[StringOf(key)] = val
			return true
		})
		return true
	})
}

type calls struct {
	calls Value
}

func (c2 *calls) Validate(c *ValidateCtx, value interface{}) {

	c2.calls.Get(value)
}

var callVad NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {
	v, err := parseValue(i)
	if err != nil {
		return nil, fmt.Errorf("parse call error:%s path:%s", err.Error(), path)
	}
	return &calls{
		calls: v,
	}, nil
}
