package jsonschema

import (
	"fmt"
)

const (
	keyCase    = "case"
	keyDefault = "defaults"
)

type AnyOf []Validator

func (a AnyOf) Validate(c *ValidateCtx, value interface{}) {
	allErrs := []Error{}
	for _, validator := range a {
		cb := c.Clone()
		validator.Validate(cb, value)
		if len(cb.errors) == 0 {
			return
		}
		allErrs = append(allErrs, cb.errors...)
	}
	// todo 区分errors

	c.AddErrors(allErrs...)
}

func NewAnyOf(i interface{}, path string, parent Validator) (Validator, error) {
	m, ok := i.([]interface{})
	if !ok {
		return nil, fmt.Errorf("value of anyOf must be array:%v,path:%s", desc(i), path)
	}
	any := AnyOf{}
	for idx, v := range m {
		ip, err := NewProp(v, path)
		if err != nil {
			return nil, fmt.Errorf("anyOf index:%d is invalid:%w %v,path:%s", idx, err, v, path)
		}
		any = append(any, ip)
	}
	return any, nil
}

type If struct {
	Then   *Then
	Else   *Else
	v      Validator
	values map[string]interface{}
}

func (i *If) Validate(c *ValidateCtx, value interface{}) {
	cif := c.CloneWithReuse()
	defer putCtx(cif)
	i.v.Validate(cif, value)
	if len(cif.errors) == 0 {
		if i.Then != nil {
			i.Then.v.Validate(c, value)
		}
	} else {
		if i.Else != nil {
			i.Else.v.Validate(c, value)
		}
	}
}

func (i *If) Get(key string) any {
	if i.values == nil {
		return nil
	}
	return i.values[key]
}

func NewIf(i interface{}, path string, parent Validator) (Validator, error) {
	ifp, err := NewProp(i, path, func(p *ArrProp) {
		p.ctx = map[string]any{
			keyIsInIf: true,
		}
		p.parent = parent
	})
	if err != nil {
		return nil, err
	}

	iff := &If{
		v: ifp,
	}
	pp, ok := parent.(*ArrProp)
	if ok {
		then, ok := pp.Get("then").(*Then)
		if ok {
			iff.Then = then
		}
		elsef, ok := pp.Get("else").(*Else)
		if ok {
			iff.Else = elsef
		}
	}
	return iff, nil
}

type Then struct {
	v Validator
}

func (t *Then) Validate(c *ValidateCtx, value interface{}) {
	// then 不能主动调用
}

type Else struct {
	v Validator
}

func (e *Else) Validate(c *ValidateCtx, value interface{}) {
	//panic("implement me")
}

func NewThen(i interface{}, path string, parent Validator) (Validator, error) {
	v, err := NewProp(i, path)
	if err != nil {
		return nil, err
	}
	return &Then{
		v: v,
	}, nil
}

func NewElse(i interface{}, path string, parent Validator) (Validator, error) {
	v, err := NewProp(i, path)
	if err != nil {
		return nil, err
	}
	return &Else{
		v: v,
	}, nil
}

type Not struct {
	v    Validator
	Path string
}

func (n Not) Validate(c *ValidateCtx, value interface{}) {
	cn := c.CloneWithReuse()
	defer putCtx(cn)
	n.v.Validate(cn, value)
	//fmt.Println(ners,value)
	if len(cn.errors) == 0 {
		c.AddErrors(Error{
			Path: n.Path,
			Info: "is not valid",
		})
	}
}

func NewNot(i interface{}, path string, parent Validator) (Validator, error) {
	p, err := NewProp(i, path, func(p *ArrProp) {
		p.parent = parent
	})
	if err != nil {
		return nil, err
	}
	return Not{v: p}, nil
}

type AllOf []Validator

func (a AllOf) Validate(c *ValidateCtx, value interface{}) {
	for _, validator := range a {
		validator.Validate(c, value)

	}
}

func NewAllOf(i interface{}, path string, parent Validator) (Validator, error) {
	arr, ok := i.([]interface{})
	if !ok {
		return nil, fmt.Errorf("value of 'allOf' must be array: %v", desc(i))
	}
	all := AllOf{}
	for _, ai := range arr {
		iv, err := NewProp(ai, path, func(p *ArrProp) {
			p.parent = parent
		})
		if err != nil {
			return nil, err
		}
		all = append(all, iv)
	}
	return all, nil
}

type Dependencies struct {
	Val  map[string][]string
	Path string
}

func (d *Dependencies) Validate(c *ValidateCtx, value interface{}) {
	m, ok := value.(map[string]interface{})
	if !ok {
		return
	}
	// 如果存在key，那么必须存在某些key
	for key, vals := range d.Val {
		_, ok := m[key]
		if ok {
			for _, val := range vals {
				_, ok = m[val]
				if !ok {
					c.AddErrors(Error{
						Path: appendString(d.Path, ".", val),
						Info: "is required",
					})
				}
			}
		}
	}
}

func NewDependencies(i interface{}, path string, parent Validator) (Validator, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("value of dependencies must be map[string][]string :%v", desc(i))
	}
	vad := &Dependencies{
		Val:  map[string][]string{},
		Path: path,
	}
	for key, arris := range m {
		arrs, ok := arris.([]interface{})
		if !ok {
			return nil, fmt.Errorf("value of dependencies must be map[string][]string :%v,path:%s", desc(i), path)
		}
		strs := make([]string, len(arrs))
		for idx, item := range arrs {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("value of dependencies must be map[string][]string :%v,path:%s", desc(i), path)

			}
			strs[idx] = str
		}
		vad.Val[key] = strs

	}
	return vad, nil
}

/*
{
	"keyMatch":{
		"key1":"biaoge"
	}
}
*/

type KeyMatch struct {
	//Val    map[*JsonPathCompiled]Value
	Val    sliceMap[*JsonPathCompiled, Value]
	Path   string
	isInIf bool
}

func (k *KeyMatch) Validate(c *ValidateCtx, value interface{}) {
	mm, ok := value.(map[string]interface{})
	if !ok {
		//c.AddError(Error{
		//	Path: k.Path,
		//	Info: "value is not object",
		//})
		return
	}
	k.Val.Range(func(key *JsonPathCompiled, want Value) bool {
		target, _ := key.Get(value)
		//target := m[key]
		ww := want.Get(mm)

		switch ww.(type) {
		case string:
			if StringOf(ww) == StringOf(target) {
				return true
			}
		case bool:
			if BoolOf(ww) == BoolOf(target) {
				return true
			}
		}

		if target != ww {
			if k.isInIf {
				// if 中的error 不需要返回出来，只需要有error 即可
				c.AddError(Error{})
			} else {
				c.AddError(Error{
					Path: appendString(k.Path, ".", key.rawPath),
					Info: fmt.Sprintf("value must be %v", ww),
				})
			}

		}
		return true
	})
	//for key, want := range k.Val {
	//
	//}
}

func NewKeyMatch(i interface{}, path string, parent Validator) (Validator, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(" %s value of keyMatch must be map[string]interface{} :%v", path, desc(i))

	}

	vm := sliceMap[*JsonPathCompiled, Value]{}
	for key, val := range m {
		jp, err := parseJpathCompiled(key)
		if err != nil {
			return nil, fmt.Errorf("%s key of keyMatch must valid jsonpath %v %v ", path, desc(i), path)
		}

		v, err := parseValue(val)
		if err != nil {
			return nil, fmt.Errorf("%s value of keyMatch must valid value %v %v ", path, desc(i), path)
		}
		//vm[jp] = v
		vm.Set(jp, v)
	}

	return &KeyMatch{
		Val:    vm,
		Path:   path,
		isInIf: isInIf(parent),
	}, nil
}

var (
	keyIsInIf = "in_if"
)

func isInIf(parent Validator) bool {
	v, ok := parent.(valuer)
	if !ok {
		return false
	}
	is, _ := v.GetVal(keyIsInIf).(bool)
	return is
}

/*
	{
		"switch":"tsy",
		"cases":{
			"key1":{},
			"key2":{}
		},
		"default":{}
	}
*/
type Switch struct {
	Switch  string
	Case    map[string]Validator
	Default *Default
}

func (s *Switch) Validate(c *ValidateCtx, value interface{}) {
	m, ok := value.(map[string]interface{})
	if !ok {
		if s.Default != nil {
			s.Default.p.Validate(c, value)
		}
		return
	}
	for cas, validator := range s.Case {
		if cas == StringOf(m[s.Switch]) {
			validator.Validate(c, value)
			return
		}
	}
	if s.Default != nil {
		s.Default.p.Validate(c, value)
	}
}

func NewSwitch(i interface{}, path string, parent Validator) (Validator, error) {
	key, ok := i.(string)
	if !ok {
		return nil, fmt.Errorf("value of switch must be string path:%s", path)
	}

	s := &Switch{
		Switch: key,
		Case:   map[string]Validator{},
	}
	ap, ok := parent.(*ArrProp)
	if !ok {
		return s, nil
	}
	cases, ok := ap.Get(keyCase).(Cases)
	if ok {
		s.Case = cases
	}
	def, ok := ap.Get(keyDefault).(*Default)
	if ok {
		s.Default = def
	}
	return s, nil
}

type Default struct {
	p *ArrProp
}

func (d Default) Validate(c *ValidateCtx, value interface{}) {
}

func NewDefault(i interface{}, path string, parent Validator) (Validator, error) {
	da, err := NewProp(i, path)
	if err != nil {
		return nil, err
	}
	return &Default{p: da.(*ArrProp)}, nil
}

type Cases map[string]Validator

func (c2 Cases) Validate(c *ValidateCtx, value interface{}) {

}

func NewCases(i interface{}, path string, parent Validator) (Validator, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("value of case must be map,path: %s", path)
	}
	cases := make(Cases)
	for key, val := range m {
		vad, err := NewProp(val, path)
		if err != nil {
			return nil, err
		}
		cases[key] = vad
	}
	return cases, nil
}
