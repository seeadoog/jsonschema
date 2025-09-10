package expr

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

var (
	funtables = map[string]ScriptFunc{
		"append":         appendFunc,
		"join":           joinFunc,
		"eq":             eqFunc,
		"eqs":            eqsFunc,
		"neq":            notEqFunc,
		"lt":             lessFunc,
		"lte":            lessOrEqual,
		"gt":             largeFunc,
		"gte":            largeOrEqual,
		"neqs":           notEqSFunc,
		"not":            notFunc,
		"or":             orFunc,
		"and":            andFunc,
		"len":            lenFunc,
		"in":             inFunc,
		"print":          printFunc,
		"add":            addFunc,
		"sub":            subFunc,
		"mul":            mulFunc,
		"mod":            modFunc,
		"div":            divFunc,
		"delete":         deleteFunc,
		"get":            getFunc,
		"set":            setFunc,
		"set_index":      setIndex,
		"str.has_prefix": hasPrefixFunc,
		"str.has_suffix": hasSuffixFunc,
		"str.join":       joinFunc,
		"str.split":      splitFunc,
		"str.to_upper":   toUpperFunc,
		"str.to_lower":   toLowerFunc,
		"str.trim":       trimFunc,
		"json.to":        jsonEncode,
		"json.from":      jsonDecode,
		"time.now":       timeNow,
		"time.format":    timeFormat,
		"type":           typeOfFunc,
		"slice.new":      newArrFunc,
		"slice.cut":      arrSliceFunc,
		"ternary":        ternaryFunc,
		"string":         stringFunc,
		"number":         numberFunc,
		"bool":           boolFunc,
		"bytes":          bytesFuncs,
		"base64.encode":  base64Encode,
		"base64.decode":  base64Decode,
		"md5":            md5SumFunc,
		"sha256":         sha256Func,
		"hmac.sha256":    hmacSha266Func,
		"hex.encode":     hexEncodeFunc,
		"hex.decode":     hexDecodeFunc,
		"sprintf":        sprintfFunc,
		"http.request":   httpRequest,
	}
)

func RegisterDynamicFunc(funName string) {
	funtables[funName] = func(ctx *Context, args ...Val) any {
		if ctx.funcs == nil {
			return nil
		}
		f := ctx.funcs[funName]
		if f == nil {
			return nil
		}
		return f(ctx, args...)
	}
}

func RegisterFunc(funName string, f ScriptFunc) {
	funtables[funName] = f
}

var appendFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) == 0 {
		return nil
	}
	a1 := args[0].Val(ctx)
	switch a1 := a1.(type) {
	case string:
		sb := strings.Builder{}
		sb.WriteString(a1)
		for _, v := range args[1:] {
			sb.WriteString(StringOf(v.Val(ctx)))
		}
		return sb.String()
	case []byte:

	case []any:
		for _, v := range args[1:] {
			a1 = append(a1, v.Val(ctx))
		}

		return a1
	}

	return nil
}

var joinFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) != 2 {
		return ""
	}

	arg := args[0].Val(ctx)
	sep := StringOf(args[1].Val(ctx))

	length := 0
	var index func(i int) string
	switch arg := arg.(type) {
	case []any:
		length = len(arg)
		index = func(i int) string {
			return StringOf(arg[i])
		}
	case []string:
		length = len(arg)
		index = func(i int) string {
			return arg[i]
		}
	default:
		return ""
	}
	switch length {
	case 0:
		return ""
	case 1:
		return index(0)
	}
	sb := strings.Builder{}
	sb.WriteString(index(0))
	for i := 1; i < length; i++ {
		sb.WriteString(sep)
		sb.WriteString(index(i))
	}
	return sb.String()
}

var eqFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) != 2 {
		return false
	}
	return args[0].Val(ctx) == args[1].Val(ctx)
}
var eqsFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) != 2 {
		return false
	}
	a0 := args[0].Val(ctx)
	a1 := args[1].Val(ctx)
	if a0 == a1 {
		return true
	}
	return StringOf(a0) == StringOf(a1)
}

var orFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	for _, arg := range args {
		v := arg.Val(ctx)
		if v != nil {

			switch vb := v.(type) {
			case bool:
				if vb {
					return true
				}
				continue
			case float64:
				if vb == 0 {
					continue
				}
			case int:
				if vb == 0 {
					continue
				}
			case string:
				if vb == "" {
					continue
				}
			}

			return v
		}
	}
	return nil
}

var andFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	for _, arg := range args {
		if !BoolOf(arg.Val(ctx)) {
			return false
		}
	}
	return true
}

var printFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	argv := make([]any, 0, len(args))
	for _, arg := range args {
		argv = append(argv, arg.Val(ctx))
	}
	fmt.Println(argv...)
	return nil
}

var sprintfFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) < 1 {
		return ""
	}
	format := StringOf(args[0].Val(ctx))
	vals := make([]any, 0, len(args)-1)
	for _, arg := range args[1:] {
		vals = append(vals, arg.Val(ctx))
	}
	return fmt.Sprintf(format, vals...)
}

func FuncDefine1[A1 any, R any](f func(a A1) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		if len(args) != 1 {
			return nil
		}
		a, _ := args[0].Val(ctx).(A1)
		return f(a)
	}
}
func FuncDefine[R any](f func() R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		return f()
	}
}

func FuncDefine2[A1 any, A2 any, R any](f func(a A1, b A2) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		if len(args) != 2 {
			return nil
		}
		a, _ := args[0].Val(ctx).(A1)

		b, _ := args[1].Val(ctx).(A2)

		return f(a, b)
	}
}

func FuncDefine3[A1 any, A2 any, A3 any, R any](f func(a A1, b A2, c A3) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		if len(args) != 3 {
			return nil
		}
		a, _ := args[0].Val(ctx).(A1)

		b, _ := args[1].Val(ctx).(A2)

		c, _ := args[2].Val(ctx).(A3)

		return f(a, b, c)
	}
}
func FuncDefine4[A1 any, A2 any, A3 any, A4 any, R any](f func(a A1, b A2, c A3, d A4) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		if len(args) != 4 {
			return nil
		}
		a, _ := args[0].Val(ctx).(A1)

		b, _ := args[1].Val(ctx).(A2)

		c, _ := args[2].Val(ctx).(A3)
		d, _ := args[3].Val(ctx).(A4)

		return f(a, b, c, d)
	}
}

func FuncDefine5[A1 any, A2 any, A3 any, A4 any, A5 any, R any](f func(a A1, b A2, c A3, d A4, e A5) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		if len(args) != 5 {
			return nil
		}
		a, _ := args[0].Val(ctx).(A1)

		b, _ := args[1].Val(ctx).(A2)

		c, _ := args[2].Val(ctx).(A3)
		d, _ := args[3].Val(ctx).(A4)
		e, _ := args[4].Val(ctx).(A5)

		return f(a, b, c, d, e)
	}
}

func FuncDefine1WithDef[A any, R any](f func(a A) R, def R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		if len(args) != 1 {
			return def
		}
		a, ok := args[0].Val(ctx).(A)
		if !ok {
			return def
		}
		return f(a)
	}
}

func FuncDefineN[T any, R any](f func(a ...T) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {

		argv := make([]T, 0, len(args))
		for _, arg := range args {
			arg, ok := arg.Val(ctx).(T)
			if !ok {
				return nil
			}
			argv = append(argv, arg)
		}
		return f(argv...)
	}
}

var setFunc = FuncDefine3(func(m map[string]any, b string, c any) any {
	m[b] = c
	return nil
})

var setIndex = FuncDefine3(func(m []any, b float64, c any) any {
	m[int(b)] = c
	return nil
})

var deleteFunc = FuncDefine2(func(m map[string]any, b string) any {
	delete(m, b)
	return nil
})

var getFunc = FuncDefine2(func(m map[string]any, b string) any {
	return m[b]
})

var addFunc = FuncDefineN(func(a ...float64) any {
	sum := 0.0
	for _, v := range a {
		sum += v
	}
	return sum
})

var subFunc = FuncDefine2(func(a, b float64) any {
	return a - b
})
var divFunc = FuncDefine2(func(a, b float64) any {
	return a / b
})
var mulFunc = FuncDefine2(func(a, b float64) any {
	return a * b
})
var modFunc = FuncDefine2(func(a, b float64) any {
	return float64(int(a) % int(b))
})
var hasPrefixFunc = FuncDefine2(func(a, b string) any {
	return strings.HasPrefix(a, b)
})

var hasSuffixFunc = FuncDefine2(func(a, b string) any {
	return strings.HasSuffix(a, b)
})

var trimFunc = FuncDefine1(strings.TrimSpace)
var splitFunc = FuncDefine2(func(a, b string) any {
	vals := strings.Split(a, b)
	va := make([]any, 0, len(vals))
	for _, v := range vals {
		va = append(va, v)
	}
	return va
})

var toUpperFunc = FuncDefine1(strings.ToUpper)

var toLowerFunc = FuncDefine1(strings.ToLower)

var jsonEncode = FuncDefine1(func(a any) any {
	bs, _ := json.Marshal(a)
	return unsafe.String(unsafe.SliceData(bs), len(bs))
})

var jsonDecode = FuncDefine1(func(a any) (res any) {
	switch a := a.(type) {
	case string:
		json.Unmarshal(unsafe.Slice(unsafe.StringData(a), len(a)), &res)
	case []byte:
		json.Unmarshal(a, &res)
	}
	return res
})

var timeFormat = FuncDefine2(func(tim time.Time, format string) any {
	return tim.Format(format)
})

var timeNow = FuncDefine(time.Now)

var notFunc = FuncDefine1(func(a any) bool {
	return !BoolOf(a)
})

var notEqFunc = FuncDefine2(func(a any, b any) bool {
	return a != b
})
var notEqSFunc = FuncDefine2(func(a any, b any) bool {
	if a != b {
		return true
	}
	return StringOf(a) != StringOf(b)
})

var largeFunc = FuncDefine2(func(a float64, b float64) bool {
	return a > b
})
var largeOrEqual = FuncDefine2(func(a float64, b float64) bool {
	return a >= b
})
var lessFunc = FuncDefine2(func(a float64, b float64) bool {
	return a < b
})
var lessOrEqual = FuncDefine2(func(a float64, b float64) bool {
	return a <= b
})

var typeOfFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) != 1 {
		return nil
	}
	a := args[0].Val(ctx)
	switch a.(type) {
	case string:
		return "string"
	case []byte:
		return "bytes"
	case float64, int:
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "nil"
	case []any:
		return "array"
	default:
		return reflect.TypeOf(a).String()
	}
}

var newArrFunc = FuncDefine1(func(a float64) any {
	return make([]any, 0, int(a))
})

var arrSliceFunc = FuncDefine3(func(arr []any, start, end float64) any {
	endi := int(end)
	starti := int(start)
	if len(arr) < endi {
		endi = len(arr)
	}
	return arr[starti:endi]
})

var ternaryFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) != 3 {
		return nil
	}
	ok := BoolOf(args[0].Val(ctx))
	if ok {
		return args[1].Val(ctx)
	}
	return args[2].Val(ctx)
}

var stringFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) != 1 {
		return nil
	}
	return StringOf(args[0].Val(ctx))
}

var boolFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) != 1 {
		return nil
	}
	return BoolOf(args[0].Val(ctx))
}

var numberFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) != 1 {
		return nil
	}
	return NumberOf(args[0].Val(ctx))
}

//var newStringBuilder ScriptFunc = func(ctx *Context, args ...Val) any {
//	return &strings.Builder{}
//}
//
//var stringBuilderWrite ScriptFunc = func(ctx *Context, args ...Val) any {
//	if len(args) <= 0 {
//		return nil
//	}
//	sb := &strings.Builder{}
//}

var base64Encode = FuncDefine1WithDef(func(a any) string {
	switch v := a.(type) {
	case string:
		return base64.StdEncoding.EncodeToString(ToBytes(v))
	case []byte:
		return base64.StdEncoding.EncodeToString(v)
	default:
		return ""
	}
}, "")

var base64Decode = FuncDefine1WithDef(func(a string) []byte {
	bs, _ := base64.StdEncoding.DecodeString(a)
	return bs
}, nil)

var bytesFuncs = FuncDefine1(BytesOf)

var md5SumFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	var data []byte
	if len(args) > 0 {
		data = BytesOf(args[0].Val(ctx))
	}
	res := md5.Sum(data)
	return res[:]
}

var sha256Func ScriptFunc = func(ctx *Context, args ...Val) any {
	var data []byte
	if len(args) > 0 {
		data = BytesOf(args[0].Val(ctx))
	}
	res := sha256.New()
	res.Write(data)
	return res.Sum(nil)
}

var hexEncodeFunc = FuncDefine1WithDef(func(a any) string {
	return hex.EncodeToString(BytesOf(a))
}, "")

var hexDecodeFunc = FuncDefine1WithDef(func(a any) []byte {
	data, _ := hex.DecodeString(StringOf(a))
	return data
}, nil)

var hmacSha266Func = FuncDefine2(func(a any, b any) []byte {
	h := hmac.New(sha256.New, BytesOf(b))
	h.Write(BytesOf(a))
	return h.Sum(nil)
})
var lenFunc = FuncDefine1(func(a any) float64 {
	switch a := a.(type) {
	case string:
		return float64(len(a))
	case []byte:
		return float64(len(a))
	case []any:
		return float64(len(a))
	case map[string]interface{}:
		return float64(len(a))
	case []string:
		return float64(len(a))
	default:
		return 0
	}
})
var inFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) < 2 {
		return false
	}
	arg := args[0].Val(ctx)
	targets := args[1:]
	for _, target := range targets {
		if arg == target.Val(ctx) {
			return true
		}
	}
	return false
}
