package expr

import (
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
	var sv S
	RegisterObjFunc(TypeOf(sv), name, fn, 1)
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
	var sv S
	RegisterObjFunc(TypeOf(sv), name, fn, 1)
}

func SelfDefine0[S any, R any](name string, f func(ctx *Context, self S) R) {

	fn := func(ctx *Context, self any, args ...Val) any {

		sl, _ := self.(S)
		return f(ctx, sl)
	}
	var sv S
	RegisterObjFunc(TypeOf(sv), name, fn, 0)

}

type objectFunc struct {
	argsNum int
	name    string
	fun     SelfFunc
}

var objFuncMap = map[Type]map[string]*objectFunc{}

func RegisterObjFunc(ty Type, name string, fun SelfFunc, argsNum int) {
	fm := objFuncMap[ty]
	if fm == nil {
		fm = map[string]*objectFunc{}
		objFuncMap[ty] = fm
	}
	fm[name] = &objectFunc{argsNum, name, fun}
}

type objFuncVal struct {
	funcName string
	args     []Val
}

func (o *objFuncVal) Val(c *Context) any {
	return nil
}

func init() {
	SelfDefine1("write", func(ctx *Context, self *strings.Builder, str any) any {
		self.WriteString(StringOf(str))
		return self
	})
	SelfDefine0("string", func(ctx *Context, self *strings.Builder) any {
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

	SelfDefine0("bytes", func(ctx *Context, self []byte) []byte {
		return self
	})
	SelfDefine0("string", func(ctx *Context, self string) string {
		return self
	})
	SelfDefine0("bytes", func(ctx *Context, self string) []byte {
		return ToBytes(self)
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
}
