package expr

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type Type unsafe.Pointer

type typeI struct {
	T, D unsafe.Pointer
}

func TypeOf(v interface{}) Type {
	return Type((*typeI)(unsafe.Pointer(&v)).T)
}

type SelfFunc func(ctx *Context, self any, args ...Val) any

func SelfDefine1[A any, S any, R any](name string, f func(ctx *Context, self S, a A) R) {
	fn := func(ctx *Context, self any, args ...Val) any {
		if len(args) != 1 {
			return newErrorf("func %s expects 1 arg, got %d", name, len(args))
		}
		a, _ := args[0].Val(ctx).(A)
		sl, _ := self.(S)
		return f(ctx, sl, a)
	}
	doc := fmt.Sprintf("%s( %v)%v", name, typeOf[A](), typeOf[R]())
	RegisterObjFunc[S](name, fn, 1, doc)
}

func typeOf[T any]() string {
	var t T
	s := reflect.TypeOf(t)
	if s == nil {
		return "any"
	}

	return s.String()
}

func SelfDefine2[A, B any, S any, R any](name string, f func(ctx *Context, self S, a A, b B) R) {
	fn := func(ctx *Context, self any, args ...Val) any {
		if len(args) != 2 {
			return newErrorf("func %s expects 1 arg, got %d", name, len(args))
		}
		a, _ := args[0].Val(ctx).(A)
		b, _ := args[1].Val(ctx).(B)
		sl, _ := self.(S)
		return f(ctx, sl, a, b)
	}
	RegisterObjFunc[S](name, fn, 2, fmt.Sprintf("%s( %v, %v)%v", name, typeOf[A](), typeOf[B](), typeOf[R]()))
}

func SelfDefine0[S any, R any](name string, f func(ctx *Context, self S) R) {

	fn := func(ctx *Context, self any, args ...Val) any {

		sl, _ := self.(S)
		return f(ctx, sl)
	}
	RegisterObjFunc[S](name, fn, 0, fmt.Sprintf("%s()%v", name, typeOf[R]()))

}

func SelfDefineN[S any, R any](name string, f SelfFunc) {

	RegisterObjFunc[S](name, f, -1, fmt.Sprintf("%s()%v", name, typeOf[R]()))

}

type objectFunc struct {
	typeI   string
	argsNum int
	name    string
	fun     SelfFunc
	doc     string
}

var objFuncMap = map[Type]map[string]*objectFunc{}

func RegisterObjFunc[T any](name string, fun SelfFunc, argsNum int, doc string) {
	var o T
	ty := TypeOf(o)
	rt := reflect.TypeOf(o)
	fm := objFuncMap[ty]
	if fm == nil {
		fm = map[string]*objectFunc{}
		objFuncMap[ty] = fm
	}
	fm[name] = &objectFunc{rt.String(), argsNum, name, fun, doc}
}

type objFuncVal struct {
	funcName string
	args     []Val
}

func (o *objFuncVal) Val(c *Context) any {
	return nil
}

func init() {
	//SelfDefine1("write", func(ctx *Context, self *strings.Builder, str any) *strings.Builder {
	//	self.WriteString(StringOf(str))
	//	return self
	//})
	SelfDefineN[*strings.Builder, *strings.Builder]("write", func(ctx *Context, self any, args ...Val) any {
		sb := self.(*strings.Builder)
		for _, arg := range args {
			sb.WriteString(StringOf(arg.Val(ctx)))
		}
		return sb
	})
	SelfDefine0("string", func(ctx *Context, self *strings.Builder) string {
		return self.String()
	})
	RegisterFunc("str.builder", func(ctx *Context, args ...Val) any {
		return new(strings.Builder)
	}, 0)
	SelfDefine1("format", func(ctx *Context, self time.Time, fmt string) string {
		return self.Format(fmt)
	})
	SelfDefine0("unix", func(ctx *Context, self time.Time) float64 {
		return float64(self.Unix())
	})
	SelfDefine0("unix_nano", func(ctx *Context, self time.Time) float64 {
		return float64(self.UnixNano())
	})

	SelfDefine1("has_prefix", func(ctx *Context, self string, str string) bool {
		return strings.HasPrefix(self, str)
	})
	SelfDefine1("has_suffix", func(ctx *Context, self string, str string) bool {
		return strings.HasSuffix(self, str)
	})

	SelfDefine0("trim_space", func(ctx *Context, self string) string {
		return strings.TrimSpace(self)
	})

	SelfDefine1("trim", func(ctx *Context, self string, cutset string) string {
		return strings.Trim(self, cutset)
	})

	SelfDefine1("trim_left", func(ctx *Context, self string, cutset string) string {
		return strings.TrimLeft(self, cutset)
	})
	SelfDefine1("trim_right", func(ctx *Context, self string, cutset string) string {
		return strings.TrimRight(self, cutset)
	})

	SelfDefine2("slice", func(ctx *Context, self []any, a, b float64) any {
		aa := int(a)
		bb := int(b)
		if len(self) < bb || aa > bb || aa < 0 {
			return nil
		}
		return self[aa:bb]
	})
	SelfDefine2("slice", func(ctx *Context, self string, a, b float64) string {
		aa := int(a)
		bb := int(b)
		if len(self) < bb || aa > bb || aa < 0 {
			return ""
		}
		return self[aa:bb]
	})
	SelfDefine2("slice", func(ctx *Context, self []byte, a, b float64) []byte {
		aa := int(a)
		bb := int(b)
		if len(self) < bb || aa > bb || aa < 0 {
			return nil
		}
		return self[aa:bb]
	})

	SelfDefine0("len", func(ctx *Context, self string) float64 {
		return float64(len(self))
	})
	SelfDefine0("len", func(ctx *Context, self []any) float64 {
		return float64(len(self))
	})

	SelfDefine0("string", func(ctx *Context, self []byte) string {
		return ToString(self)
	})
	SelfDefine0("string", func(ctx *Context, self bool) string {
		return strconv.FormatBool(self)
	})

	SelfDefine0("string", func(ctx *Context, self float64) string {
		return strconv.FormatFloat(self, 'f', -1, 64)
	})

	SelfDefine0("bytes", func(ctx *Context, self []byte) []byte {
		return self
	})
	SelfDefine0("string", func(ctx *Context, self string) string {
		return self
	})
	SelfDefine0("bytes", func(ctx *Context, self string) []byte {
		return ToBytes(self)
	})
	SelfDefine1("has", func(ctx *Context, self string, s string) bool {
		return strings.Contains(self, s)
	})
	SelfDefine1("contains", func(ctx *Context, self string, s string) bool {
		return strings.Contains(self, s)
	})

	SelfDefine0("md5", func(ctx *Context, self string) []byte {
		h := md5.New()
		h.Write(ToBytes(self))

		return h.Sum(nil)
	})
	SelfDefine0("hex", func(ctx *Context, self string) string {
		return hex.EncodeToString(ToBytes(self))
	})
	SelfDefine0("hex", func(ctx *Context, self []byte) string {
		return hex.EncodeToString(self)
	})
	SelfDefine0("bytes", func(ctx *Context, self []byte) []byte {
		h := md5.New()
		h.Write(self)
		return h.Sum(nil)
	})

	SelfDefine0("copy", func(ctx *Context, b []byte) []byte {
		dst := make([]byte, len(b))
		copy(dst, b)
		return dst
	})
	SelfDefine0("base64", func(ctx *Context, b []byte) string {
		return base64EncodeToString(b)
	})

	SelfDefine0("base64", func(ctx *Context, self string) string {
		return base64EncodeToString(ToBytes(self))
	})

	SelfDefine0("base64d", func(ctx *Context, b []byte) []byte {
		d, _ := base64DecodeString(ToString(b))
		return d
	})
	SelfDefine0("base64d", func(ctx *Context, self string) []byte {
		d, _ := base64DecodeString(self)
		return d
	})

	SelfDefine0("type", func(ctx *Context, self string) string {
		return "string"
	})
	SelfDefine0("type", func(ctx *Context, self float64) string {
		return "number"
	})
	SelfDefine0("type", func(ctx *Context, self bool) string {
		return "boolean"
	})
	SelfDefine0("type", func(ctx *Context, self []byte) string {
		return "bytes"
	})
	objFuncMap[TypeOf(nil)] = map[string]*objectFunc{
		"type": {
			typeI:   "nil",
			argsNum: 0,
			name:    "type",
			fun: func(ctx *Context, self any, args ...Val) any {
				return "nil"
			},
			doc: "type()string",
		},
		"string": {
			typeI:   "nil",
			argsNum: 0,
			name:    "string",
			fun: func(ctx *Context, self any, args ...Val) any {
				return ""
			},
			doc: "string()string",
		},
		"number": {
			typeI:   "nil",
			argsNum: 0,
			name:    "number",
			fun: func(ctx *Context, self any, args ...Val) any {
				return 0.0
			},
			doc: "number()float64",
		},
		"boolean": {
			typeI:   "nil",
			argsNum: 0,
			name:    "boolean",
			fun: func(ctx *Context, self any, args ...Val) any {
				return false
			},
			doc: "bool()bool",
		},
	}
	SelfDefine2("set", func(ctx *Context, self map[string]any, a string, b any) map[string]any {
		self[a] = b
		return self
	})
	SelfDefine1("get", func(ctx *Context, self map[string]any, a string) any {
		return self[a]
	})
	SelfDefine0("len", func(ctx *Context, self map[string]any) float64 {
		return float64(len(self))
	})
	SelfDefine1("delete", func(ctx *Context, self map[string]any, a string) map[string]any {
		delete(self, a)
		return self
	})
	SelfDefine1("get", func(ctx *Context, self []any, a float64) any {
		n := int(a)
		if n >= len(self) {
			return nil
		}
		return self[n]
	})

	SelfDefine1("sub", func(ctx *Context, self time.Time, tm time.Time) float64 {
		return float64(self.Sub(tm) / 1e6)
	})
	SelfDefine1("add_mill", func(ctx *Context, self time.Time, mill float64) time.Time {
		return self.Add(time.Duration(mill * 1e6))
	})
	SelfDefine0("day", func(ctx *Context, self time.Time) float64 {
		return float64(self.Day())
	})
	SelfDefine0("hour", func(ctx *Context, self time.Time) float64 {
		return float64(self.Hour())
	})
	SelfDefine0("month", func(ctx *Context, self time.Time) float64 {
		return float64(self.Month())
	})
	SelfDefine0("year", func(ctx *Context, self time.Time) float64 {
		return float64(self.Year())
	})
	SelfDefine0("utc", func(ctx *Context, self time.Time) time.Time {
		return self.UTC()
	})
	SelfDefine0("local", func(ctx *Context, self time.Time) time.Time {
		return self.Local()
	})

	RegisterFunc("regexp.new", FuncDefine1(func(a string) any {
		reg, err := regexp.Compile(a)
		if err != nil {
			return nil
		}
		return reg
	}), 1)
	SelfDefine1("match", func(ctx *Context, self *regexp.Regexp, src string) bool {
		return self.MatchString(src)
	})

	RegisterFunc("url.new_values", FuncDefine(func() any {
		uv := url.Values{}
		return uv
	}), 0)

	SelfDefine1("get", func(ctx *Context, self url.Values, key string) string {
		return self.Get(key)
	})
	SelfDefine2("set", func(ctx *Context, self url.Values, key string, val any) any {
		self.Set(key, StringOf(val))
		return self
	})
	SelfDefine0("encode", func(ctx *Context, self url.Values) string {
		return self.Encode()
	})

	//var (
	//	arrKeys = []string{""}
	//)
	//var (
	//	mapKeys = []string{"$key", "$val"}
	//)
	RegisterObjFunc[[]any]("all", func(ctx *Context, self any, args ...Val) any {
		if len(args) != 1 {
			return newErrorf("all expects 1 arg")
		}
		dst := make([]any, 0, len(args))
		forRangeExec(args[0], ctx, self, func(k, v any, val Val) any {
			if BoolCond(val.Val(ctx)) {
				dst = append(dst, v)
			}
			return nil
		})
		return dst
	}, 1, "all(cond)[]any")
	RegisterObjFunc[[]any]("filter", func(ctx *Context, self any, args ...Val) any {
		if len(args) != 1 {
			return newErrorf("filter expects 1 arg")
		}

		dst := make([]any, 0, len(self.([]any)))
		forRangeExec(args[0], ctx, self, func(k, v any, val Val) any {
			dst = append(dst, val.Val(ctx))
			return nil
		})
		return dst
	}, 1, "all(cond)[]any")

	RegisterObjFunc[[]any]("for", func(ctx *Context, self any, args ...Val) any {

		if len(args) != 1 {
			return newErrorf("for expects 1 arg")
		}
		forRangeExec(args[0], ctx, self, func(_, _ any, val Val) any {
			v := val.Val(ctx)
			if err := convertToError(v); err != nil {
				return err
			}
			return nil
		})
		return nil
	}, 1, "for(expr)")

	RegisterObjFunc[map[string]any]("for", func(ctx *Context, self any, args ...Val) any {

		if len(args) != 1 {
			return newErrorf("for expects 1 arg")
		}
		forRangeExec(args[0], ctx, self, func(_, _ any, val Val) any {
			v := val.Val(ctx)
			if err := convertToError(v); err != nil {
				return err
			}
			return nil
		})
		return nil
	}, 1, "for(expr)")

	SelfDefine0("json_str", func(ctx *Context, self map[string]any) string {
		bs, _ := json.Marshal(self)
		return ToString(bs)
	})
	SelfDefine0("json_str", func(ctx *Context, self []any) string {
		bs, _ := json.Marshal(self)
		return ToString(bs)
	})
	SelfDefine0("json_str", func(ctx *Context, self float64) string {
		bs, _ := json.Marshal(self)
		return ToString(bs)
	})
	SelfDefine0("json_str", func(ctx *Context, self string) string {
		bs, _ := json.Marshal(self)
		return ToString(bs)
	})
}
