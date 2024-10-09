package jsonschema

import (
	"fmt"
	"github.com/tidwall/gjson"
	"net"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func init() {
	// 这些显示放在funcs 里面时，不让编译通过，透。。。
	RegisterValidator("properties", NewProperties(false))
	RegisterValidator("props", NewProperties(false))
	//RegisterValidator("flexProperties", NewProperties(true))
	RegisterValidator("items", NewItems)
	RegisterValidator("anyOf", NewAnyOf)
	RegisterValidator("or", NewAnyOf)
	RegisterValidator("if", NewIf)
	RegisterValidator("else", NewElse)
	RegisterValidator("then", NewThen)
	RegisterValidator("not", NewNot)
	RegisterValidator("allOf", NewAllOf)
	RegisterValidator("and", NewAllOf)
	RegisterValidator("dependencies", NewDependencies)
	RegisterValidator("keyMatch", NewKeyMatch)
	RegisterValidator("equals", NewKeyMatch)
	RegisterValidator("eq", NewKeyMatch)

	RegisterValidator("setVal", NewSetVal)
	RegisterValidator("setExpr", NewSetExpr)
	RegisterValidator("set", NewSetVal)
	RegisterValidator("switch", NewSwitch)
	RegisterValidator(keyCase, NewCases)
	RegisterValidator(keyDefault, NewDefault)
	RegisterValidator("formatVal", NewFormatVal)
	RegisterValidator("format", NewFormat)
	RegisterValidator("additionalProperties", NewAdditionalProperties)
	RegisterValidator("multipleOf", NewMultipleOf)
	RegisterValidator("maxB64DLen", NewMaxB64DLen)
	RegisterValidator("minB64DLen", NewMinB64DLength)
	RegisterValidator("const", NewConst)
	RegisterValidator("error", newError)
	RegisterValidator("delete", newDeleteValidator)
	RegisterValidator("children", newChildrenValidator)
	RegisterValidator("uniqueItems", newUniqueItemValidator)
	RegisterValidator("maxItems", newMaxItems)
	RegisterValidator("minItems", newMinItems)
	RegisterValidator("exclusiveMaximum", NewExclusiveMaximum)
	RegisterValidator("exclusiveMinimum", NewExclusiveMinimum)
	RegisterValidator("defaultVals", NewDefaultValues)
	RegisterValidator("foreach", newForeach)
	registerCompares()
}

func registerCompares() {
	RegisterValidator("startWith", NewCompareSingle(func(actual, def string) bool {
		return strings.HasPrefix(actual, def)
	}, " should start with "))

	RegisterValidator("endWith", NewCompareSingle(func(actual, def string) bool {
		return strings.HasSuffix(actual, def)
	}, " should end with "))
	RegisterValidator("contains", NewCompareSingle(func(actual, def string) bool {
		return strings.Contains(actual, def)
	}, " should contains "))

	RegisterValidator("startWiths", NewCompare(func(actual, def string, c Context) bool {
		return strings.HasPrefix(actual, def)
	}, "should start with "))

	RegisterValidator("endWiths", NewCompare(func(actual, def string, c Context) bool {
		return strings.HasSuffix(actual, def)
	}, "should start with "))

	RegisterValidator("containss", NewCompare(func(actual, def string, c Context) bool {
		return strings.Contains(actual, def)
	}, "should contains "))

	RegisterValidator("maxLengths", NewCompare(func(actual string, def float64, c Context) bool {
		return len(actual) <= int(def)
	}, "length should less then"))

	RegisterValidator("minLengths", NewCompare(func(actual string, def float64, c Context) bool {
		return len(actual) >= int(def)
	}, "length should larger then"))

	RegisterValidator("patterns", NewCompare(func(actual string, def *regexp.Regexp, c Context) bool {
		return def.MatchString(actual)
	}, "should match regular expression", withOptParse(func(a any) (res *regexp.Regexp, err error) {
		str, ok := a.(string)
		if !ok {
			return nil, fmt.Errorf("regexp expect string")
		}
		return regexp.Compile(str)
	})))

	RegisterValidator("setMap", NewMapOpt(func(m map[string]any, key string, val any) {
		m[key] = val
	}))
	RegisterValidator("delMap", NewMapOpt(func(m map[string]any, key string, val any) {
		delete(m, key)
	}))

	RegisterValidator("del", NewMapOpt(func(m map[string]any, key string, val any) {
		delete(m, key)
	}))

	RegisterValidator("gt", NewCompareVal(func(actual float64, def float64, c Context) bool {

		return actual > def
	}, "should greater than "))

	RegisterValidator("lt", NewCompareVal(func(actual, def float64, c Context) bool {
		return actual < def
	}, "should less than"))

	RegisterValidator("gte", NewCompareVal(func(actual, def float64, c Context) bool {
		return actual >= def
	}, "should greater or equal than "))

	RegisterValidator("lte", NewCompareVal(func(actual, def float64, c Context) bool {
		return actual <= def
	}, "should less or equal  than "))

	RegisterValidator("neq", NewCompareVal(func(actual, def any, c Context) bool {
		return actual != def
	}, "should not equal with "))

	in := NewCompare(func(actual any, def []Value, c Context) bool {
		for _, a := range def {
			if a.Get(c) == actual {
				return true
			}
		}
		return false
	}, "should be one of  ", withOptParse(func(a any) (res []Value, err error) {
		data, ok := a.([]any)
		if !ok {
			return nil, fmt.Errorf("'in' or 'notin' opt right value expect slice")
		}
		res = make([]Value, 0, len(data))
		for _, v := range data {
			v2, err := parseValue(v)
			if err != nil {
				return nil, err
			}
			res = append(res, v2)
		}
		return res, nil
	}))
	RegisterValidator("in", in)

	RegisterValidator("notin", func(i interface{}, path string, parent Validator) (Validator, error) {
		vi, err := in(i, path, parent)
		if err != nil {
			return nil, err
		}
		return &Not{v: vi, Path: path}, nil
	})

	RegisterValidator("ipIn", NewCompare(func(actual string, def []*net.IPNet, c Context) bool {
		ip := net.ParseIP(actual).To4()
		if ip == nil {
			return false
		}
		for _, v := range def {
			if v.Contains(ip) {
				return true
			}
		}
		return false
	}, " ip should be within ", withOptParse(func(a any) (res []*net.IPNet, err error) {
		arr, ok := a.([]any)
		if !ok {
			return nil, fmt.Errorf("ipIn should be slice type")
		}
		for _, ipstr := range arr {
			str := StringOf(ipstr)
			if !strings.Contains(str, "/") {
				str = str + "/32"
			}
			_, ipcird, err := net.ParseCIDR(str)
			if err != nil {
				return nil, fmt.Errorf("parse %s as ip  error %w", ipstr, err)
			}
			res = append(res, ipcird)
		}
		return res, nil
	})))

}

// 忽略的校验器
var ignoreKeys = map[string]int{
	"title":       1,
	"comment":     1,
	"$comment":    1,
	"description": 1,
	"$id":         1,
	"$schema":     1,
	"id":          1,
}

var priorities = map[string]int{
	"switch":      1,
	"if":          1,
	"required":    2,
	"properties":  1,
	"maximum":     1,
	"minimum":     1,
	"defaultVals": 3,
}

func AddIgnoreKeys(key string) {
	ignoreKeys[key] = 1
}
func RegisterValidator(name string, fun NewValidatorFunc) {
	//if funcs[name] != nil {
	//	panicf("register validator error! %s already exists", name)
	//}
	funcs[name] = fun
}

var funcs = map[string]NewValidatorFunc{
	"type": NewType,
	//"types":      NewTypes,
	"maxLength":  NewMaxLen,
	"minLength":  NewMinLen,
	"maximum":    NewMaximum,
	"minimum":    NewMinimum,
	"required":   NewRequired,
	"constVal":   NewConstVal,
	"defaultVal": NewDefaultVal,
	"default":    NewDefaultVal,
	"replaceKey": NewReplaceKey,
	"enums":      NewEnums,
	"enum":       NewEnums,
	"pattern":    NewPattern,
}

type PropItem struct {
	Key string
	Val Validator
}

type ArrProp struct {
	Val    []PropItem
	Path   string
	ctx    map[string]any
	parent Validator
}

func (a *ArrProp) GetVal(key string) any {
	if a.ctx != nil {
		if v, ok := a.ctx[key]; ok {
			return v
		}
	}

	v, ok := a.parent.(valuer)
	if !ok {
		return nil
	}
	return v.GetVal(key)
}

func (a *ArrProp) GetChild(path string) Validator {
	return a.Get(path)
}

func (a *ArrProp) Validate(c *ValidateCtx, value interface{}) {
	for _, item := range a.Val {
		if item.Val == nil {
			continue
		}
		item.Val.Validate(c, value)
	}
}
func (a *ArrProp) Get(key string) Validator {
	for _, item := range a.Val {
		if item.Key == key {
			return item.Val
		}

	}
	return nil
}

type propWrap struct {
	key      string
	val      interface{}
	priority int
}

type propOpt func(p *ArrProp)

func NewProp(i interface{}, path string, opts ...propOpt) (Validator, error) {
	p := make([]PropItem, 0)
	arr := &ArrProp{
		Val:  p,
		Path: path,
	}

	for _, f := range opts {
		f(arr)
	}
	m, ok := i.(map[string]interface{})
	if !ok {
		if _, ok := i.([]interface{}); ok {
			return NewAllOf(i, path, arr)
		}
		return nil, fmt.Errorf("cannot create prop with not object type: %v,path:%s", desc(i), path)
	}

	pwaps := make([]propWrap, 0, len(p))
	for key, val := range m {

		if funcs[key] == nil {
			if ignoreKeys[key] > 0 {
				continue
			}
			return nil, fmt.Errorf("%s is unknown validator,path=%s", key, path)
		}
		pwaps = append(pwaps, propWrap{
			key:      key,
			val:      val,
			priority: priorities[key],
		})

	}

	sort.Slice(pwaps, func(i, j int) bool {
		return pwaps[i].priority < pwaps[j].priority
	}) // 对子序列排序，优先级低的先加载，优先级高的后加载

	for _, v := range pwaps {
		key := v.key
		val := v.val
		var vad Validator
		var err error
		// items 的path 不一样，
		if key == "items" {
			vad, err = funcs[key](val, path+"[*]", arr)
		} else {
			vad, err = funcs[key](val, path, arr)
		}

		if err != nil {
			return nil, fmt.Errorf("create prop error:key=%s,err=%w", key, err)
		}
		//p[key] = vad
		arr.Val = append(arr.Val, PropItem{Key: key, Val: vad})
	}
	return arr, nil
}

type Properties struct {
	properties map[string]Validator
	//properties sliceMap[string, Validator]
	//constVals            map[string]*ConstVal
	constVals            sliceMap[string, *ConstVal]
	defaultVals          sliceMap[string, *DefaultVal]
	replaceKeys          sliceMap[string, ReplaceKey]
	formats              sliceMap[string, FormatVal]
	Path                 string
	EnableUnknownField   bool
	additionalProperties Validator
}

func (p *Properties) GetChild(path string) Validator {
	return p.properties[path]
}

func (p *Properties) GValidate(ctx *ValidateCtx, val *gjson.Result) {
	//TODO implement me
	if val.Type == gjson.Null {
		return
	}
	if !val.IsObject() {
		ctx.AddError(Error{
			Path: p.Path,
			Info: "type should be object",
		})
		return
	}
	val.ForEach(func(key, value gjson.Result) bool {
		vad := p.properties[key.Str]
		if vad == nil {
			if !p.EnableUnknownField {
				ctx.AddErrorInfo(p.Path+"."+key.Str, "unknown field")
				return true
			}
			return true
		}
		panic("implement me")
	})
}

func (p *Properties) Validate(c *ValidateCtx, value interface{}) {
	if value == nil {
		return
	}

	if m, ok := value.(map[string]interface{}); ok {
		for k, v := range m {
			pv := p.properties[k]
			if pv == nil {
				if !p.EnableUnknownField {
					c.AddError(Error{
						Path: appendString(p.Path, ".", k),
						Info: "unknown field",
					})
					continue
				}
				if p.additionalProperties != nil {
					cp := c.Clone()
					p.additionalProperties.Validate(cp, v)
					// for i, e := range cp.errors {
					// 	cp.errors[i].Path = e.Path + "." + k
					// }
					c.AddErrors(cp.errors...)
				}
				continue
			}
			pv.Validate(c, v)
		}
		// 执行参数转换逻辑
		//for key, val := range p.constVals {
		//	m[key] = val.Val
		//}
		p.constVals.Range(func(key string, val *ConstVal) bool {
			m[key] = val.Val
			return true
		})
		p.defaultVals.Range(func(key string, val *DefaultVal) bool {
			if _, ok := m[key]; !ok {
				m[key] = val.Val
				pv, _ := p.properties[key]
				if pv != nil {
					// 如果默认值是map 类型，需要复制下，否则会存在 并发读写的问题
					pv.Validate(c.Clone(), copyValue(val.Val))
				}
			}
			return true
		})
		//for key, val := range p.defaultVals {
		//
		//}

		p.replaceKeys.Range(func(key string, rpk ReplaceKey) bool {
			if mv, ok := m[key]; ok {
				_, exist := m[string(rpk)]
				if exist { // 如果要替换的key 已经存在，不替换
					return true
				}
				m[string(rpk)] = mv

			}
			return true
		})
		//for key, rpk := range p.replaceKeys {
		//
		//}
		p.formats.Range(func(key string, v FormatVal) bool {
			vv, ok := m[key]
			if ok {
				m[key] = v.Convert(vv)
			}
			return true
		})
		//if len(p.formats) > 0 {
		//
		//}
	} else {
		rv := reflect.ValueOf(value)
		p.validateStruct(c, rv)

	}
}

func (p *Properties) validateStruct(c *ValidateCtx, rv reflect.Value) {
	//fmt.Println("vadd:", rv.Type().String())
	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			return
		}
		rv = rv.Elem()
		p.validateStruct(c, rv)
		return
	case reflect.Struct:
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			ft := rt.Field(i)
			propName := ft.Tag.Get("json")
			if propName == "" {
				propName = ft.Name
			}
			//fmt.Println("valds",propName)
			vad := p.properties[propName]
			if vad == nil {
				continue
			}
			fv := rv.Field(i)
			//fmt.Println("fv.", fv.String(), fv.CanInterface(), vad)
			if fv.CanInterface() {
				//vad.Validate(propName, fv.Interface(), errs)

				vad.Validate(c, fv.Interface())
			}
			// set constVal ,struct 类型无法知道目标值是否为空，无法设定默认值
			var vv interface{} = nil
			constv := p.constVals.Getv(propName)
			if constv != nil {
				vv = constv.Val
			}
			if vv == nil {
				continue
			}
			setV := reflect.ValueOf(vv)
			if setV.Kind() == fv.Kind() {
				fv.Set(setV)
			} else if setV.Kind() == reflect.Float64 {
				switch fv.Kind() {
				case reflect.Int, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Int16:
					fv.SetInt(int64(setV.Interface().(float64)))
				case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
					fv.SetUint(uint64(setV.Interface().(float64)))
				case reflect.Float32:
					fv.SetFloat(setV.Interface().(float64))
				}
			}

		}
	case reflect.Map:
		rg := rv.MapRange()
		for rg.Next() {
			key := rg.Key()
			if key.Kind() != reflect.String {
				return
			}
			val := rg.Value()
			vad := p.properties[key.String()]
			if vad != nil {
				vad.Validate(c, val.Interface())
			} else {
				if !p.EnableUnknownField {
					c.AddErrorInfo(p.Path+"."+key.String(), "unknown filed")
					return
				}
				if p.additionalProperties != nil {
					//ctx := c.Clone()
					p.additionalProperties.Validate(c, val.Interface())
					//if len(ctx.errors) > 0 {
					//for _, e := range ctx.errors {
					//	c.AddError(Error{
					//		Path: e.Path,
					//		Info: e.Info,
					//	})
					//}
					//}
				}

			}
		}
	default:
		//c.AddErrorInfo(p.Path, "invalid type , type should be object, but:%v"+rv.Type().String())
	}

}

func NewProperties(enableUnKnownFields bool) NewValidatorFunc {
	return func(i interface{}, path string, parent Validator) (validator Validator, e error) {
		m, ok := i.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot create properties with not object type: %v,flex:%v,path:%s", i, enableUnKnownFields, path)
		}
		p := &Properties{
			properties: map[string]Validator{},
			//replaceKeys:        map[string]ReplaceKey{},
			//constVals:          map[string]*ConstVal{},
			//defaultVals:        map[string]*DefaultVal{},
			//formats:            map[string]FormatVal{},
			Path:               path,
			EnableUnknownField: enableUnKnownFields,
		}
		for key, val := range m {
			vad, err := NewProp(val, appendString(path, ".", key))
			if err != nil {
				return nil, err
			}
			p.properties[key] = vad
		}
		pap, ok := parent.(*ArrProp)
		if ok {
			additional, ok := pap.Get("additionalProperties").(*AdditionalProperties)
			if ok {
				p.EnableUnknownField = additional.enableUnknownField
				p.additionalProperties = additional.validator
			}
		}
		for key, val := range p.properties {
			prop, ok := val.(*ArrProp)
			if !ok {
				continue
			}
			constVal, ok := prop.Get("constVal").(*ConstVal)
			if ok {
				p.constVals.Set(key, constVal)
			}
			defaultVal, ok := prop.Get("defaultVal").(*DefaultVal)
			if ok {
				p.defaultVals.Set(key, defaultVal)
			}

			defaultVal, ok = prop.Get("default").(*DefaultVal)
			if ok {
				p.defaultVals.Set(key, defaultVal)
			}
			replaceKey, ok := prop.Get("replaceKey").(ReplaceKey)
			if ok {
				p.replaceKeys.Set(key, replaceKey)
			}

			format, ok := prop.Get("formatVal").(FormatVal)
			if ok {
				p.formats.Set(key, format)
			}
		}

		return p, nil
	}
}

type AdditionalProperties struct {
	enableUnknownField bool
	validator          Validator
}

func (a AdditionalProperties) Validate(c *ValidateCtx, value interface{}) {

}

func NewAdditionalProperties(i interface{}, path string, parent Validator) (Validator, error) {
	//bv, ok := i.(bool)
	//if !ok {
	//	return nil, fmt.Errorf("value of 'additionalProperties' must be boolean: %v", desc(i))
	//}
	switch i := i.(type) {
	case bool:
		return &AdditionalProperties{enableUnknownField: i}, nil
	default:
		vad, err := NewProp(i, path+"{*}")
		if err != nil {
			return nil, err
		}
		return &AdditionalProperties{enableUnknownField: true, validator: vad}, nil
	}

	//return nil, fmt.Errorf("value of 'additionalProperties' must be boolean or object: %v", desc(i))
}

type AdditionalProperties2 struct {
	Validators []Validator
}

func (a *AdditionalProperties2) Validate(c *ValidateCtx, value interface{}) {

}

type errorVal struct {
	path string
	//errorInfo string
	errInfo Value
}

func (e *errorVal) Validate(c *ValidateCtx, value interface{}) {
	c.AddError(Error{
		Path: e.path,
		Info: StringOf(e.errInfo.Get(value)),
	})
}

var newError NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {
	//str, ok := i.(string)
	//if !ok{
	//	return nil,fmt.Errorf("%s error shold be string",path)
	//}
	val, err := parseValue(i)
	if err != nil {
		return nil, err
	}
	return &errorVal{
		path:    path,
		errInfo: val,
	}, nil
}

type deleteValidator struct {
	deletes []string
}

func (d *deleteValidator) Validate(c *ValidateCtx, value interface{}) {
	switch m := value.(type) {
	case map[string]interface{}:
		for _, key := range d.deletes {
			delete(m, key)
		}
	}
}

var newDeleteValidator NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {
	arr, ok := i.([]interface{})
	if !ok {
		return nil, fmt.Errorf("new delete error, value should be array")
	}
	strs := []string{}
	for _, v := range arr {
		strs = append(strs, StringOf(v))
	}
	return &deleteValidator{deletes: strs}, nil
}

type childValidator struct {
	children map[string]Validator
}

func (chd *childValidator) Validate(c *ValidateCtx, value interface{}) {
	switch v := value.(type) {
	case map[string]interface{}:
		for key, validator := range chd.children {
			val, ok := v[key]
			if ok {
				validator.Validate(c, val)
			}
		}
	}
}

var newChildrenValidator NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {
	m, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("children validator value should be map,but now is:%s", reflect.TypeOf(i).String())
	}
	chv := &childValidator{children: map[string]Validator{}}
	var err error
	for key, val := range m {
		chv.children[key], err = NewProp(val, path+"."+key)
		if err != nil {
			return nil, err
		}
	}
	return chv, nil
}

/*
children :{
	asms :{

	}
}
*/
// uniqueItems  should define together with items which restrict the array item
//to be comparable type .otherwise ,the validator may panic at runtime
// if item is not comparable type
type uniqueItems struct {
	path   string
	unique bool
}

func (u *uniqueItems) Validate(c *ValidateCtx, value interface{}) {
	if !u.unique {
		return
	}
	arr, ok := value.([]interface{})
	if !ok {
		return
	}
	okMap := make(map[interface{}]bool, len(arr))
	for _, val := range arr {
		if !isComparable(val) {
			c.AddErrorInfo(u.path, " items should be comparable type,like [ string boolean number ]")
			return
		}
		_, _exist := okMap[val]
		if _exist {
			c.AddErrorInfo(u.path, " items should be unique")
			return
		}
		okMap[val] = true
	}
}

var newUniqueItemValidator NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {
	unique, ok := i.(bool)
	if !ok {
		return nil, fmt.Errorf("%s uniqueItems value should be boolean ", path)
	}
	return &uniqueItems{unique: unique, path: path}, nil
}

func isComparable(v interface{}) bool {
	switch v.(type) {
	case string, float64, bool:
		return true
	}
	return false
}

type maxItems struct {
	val  int
	path string
}

func (m *maxItems) Validate(c *ValidateCtx, value interface{}) {
	arr, ok := value.([]interface{})
	if !ok {
		return
	}
	if len(arr) > m.val {
		c.AddErrorInfo(m.path, " max length is "+strconv.Itoa(m.val))
	}
}

var newMaxItems NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {
	val, ok := i.(float64)
	if !ok {
		return nil, fmt.Errorf("%s maxItems should be integer", path)
	}
	return &maxItems{path: path, val: int(val)}, nil
}

type minItems struct {
	val  int
	path string
}

func (m *minItems) Validate(c *ValidateCtx, value interface{}) {
	arr, ok := value.([]interface{})
	if !ok {
		return
	}
	if len(arr) < m.val {
		c.AddErrorInfo(m.path, " min length is "+strconv.Itoa(m.val))
	}
}

var newMinItems NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {
	val, ok := i.(float64)
	if !ok {
		return nil, fmt.Errorf("%s maxItems should be integer", path)
	}
	return &minItems{path: path, val: int(val)}, nil
}

func copyValue(v interface{}) interface{} {
	switch vv := v.(type) {
	case string, float64, bool:
		return v
	case map[string]interface{}:
		dst := make(map[string]interface{}, len(vv))
		for key, val := range vv {
			dst[key] = copyValue(val)
		}
		return dst
	case []interface{}:
		dst := make([]interface{}, len(vv))
		for i, val := range vv {
			dst[i] = copyValue(val)
		}
		return dst
	case nil:

		return nil
	}
	return nil
}

type exclusiveMaximum struct {
	Path   string
	Value  float64
	status int //0: 需要被校验，1 不需要校验，由 maxinum 校验，2 maxmum 也不需要校验
}

func (v *exclusiveMaximum) Validate(c *ValidateCtx, value interface{}) {
	if v.status != 0 {
		return
	}
	vv, ok := value.(float64)
	if ok {
		if vv >= float64(v.Value) {
			c.AddError(Error{Path: v.Path, Info: fmt.Sprintf("value should be < %v", v.Value)})
		}
	}
}

var NewExclusiveMaximum NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {

	switch f := i.(type) {
	case float64:
		return &exclusiveMaximum{Path: path, Value: f, status: 0}, nil
	case bool:
		status := 1
		if !f {
			status = 2
		}
		return &exclusiveMaximum{Path: path, status: status}, nil
	}
	return nil, fmt.Errorf("exclusiveMaximum should be number or bool")
}

type exclusiveMinimum struct {
	Path   string
	Value  float64
	status int
}

var NewExclusiveMinimum NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {

	switch f := i.(type) {
	case float64:
		return &exclusiveMinimum{Path: path, Value: f, status: 0}, nil
	case bool:
		status := 1
		if !f {
			status = 2
		}
		return &exclusiveMinimum{Path: path, status: status}, nil
	}
	return nil, fmt.Errorf("exclusiveMinimum should be number or bool")
}

func (v *exclusiveMinimum) Validate(c *ValidateCtx, value interface{}) {
	if v.status != 0 {
		return
	}
	vv, ok := value.(float64)
	if ok {
		if vv <= float64(v.Value) {
			c.AddError(Error{Path: v.Path, Info: fmt.Sprintf("value should be > %v", v.Value)})
		}
	}
}

func errorr[T any](f string, args ...any) (t T, err error) {
	return t, fmt.Errorf(f, args...)
}
