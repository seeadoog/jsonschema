package jsonschema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)
var(
	sprintf = fmt.Sprintf
)

type Error struct {
	Path string
	Info string
}
type ValidateCtx struct {
	errors []Error
}

func (v *ValidateCtx) AddError(e Error) {
	v.errors = append(v.errors, e)
}

func (v *ValidateCtx) AddErrors(e ...Error) {
	for i, _ := range e {
		v.AddError(e[i])
	}
}

func (v *ValidateCtx) Clone() *ValidateCtx {
	return &ValidateCtx{}
}

type Validator interface {
	Validate(c *ValidateCtx, value interface{})
}

type NewValidatorFunc func(i interface{}, path string, parent Validator) (Validator, error)

func appendString(s ...string) string {
	sb := strings.Builder{}
	for _, str := range s {
		sb.WriteString(str)
	}
	return sb.String()
}

func panicf(f string, args ...interface{}) {
	panic(fmt.Sprintf(f, args...))
}

func String(v interface{}) string {
	switch v.(type) {
	case string:
		return v.(string)
	case bool:
		if v.(bool) {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(v.(float64), 'f', -1, 64)
	case nil:
		return ""

	}
	return fmt.Sprintf("%v", v)
}

func Number(v interface{}) float64 {
	switch v.(type) {
	case float64:
		return v.(float64)
	case bool:
		if v.(bool) {
			return 1
		}
		return 0
	case string:
		i, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return i
		}
		if v.(string) == "true" {
			return 1
		}
		return 0
	}
	return 0
}

func Bool(v interface{}) bool {
	switch v.(type) {
	case float64:
		return v.(float64) > 0
	case string:
		return v.(string) == "true"
	case bool:
		return v.(bool)
	default:
		if Number(v) > 0{
			return true
		}
	}
	return false
}
func Equal(a, b interface{}) bool {
	return String(a) == String(b)
}

func desc(i interface{})string{
	ty:=reflect.TypeOf(i)
	return fmt.Sprintf("value:%v,type:%s",i,ty.String())
}
