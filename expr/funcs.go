package expr

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"
)

type innerFunc struct {
	argString string
	hasOpt    bool
	fun       ScriptFunc
	name      string
	argsNum   int
}

var (
	funtables = map[string]*innerFunc{
		"append":         {"", false, appendFunc, "append", -1},
		"join":           {"", false, joinFunc, "join", -1},
		"eq":             {"", false, eqFunc, "eq", 2},
		"eqs":            {"", false, eqsFunc, "eqs", 2},
		"neq":            {"", false, notEqFunc, "neq", 2},
		"lt":             {"", false, lessFunc, "lt", 2},
		"lte":            {"", false, lessOrEqual, "lte", 2},
		"gt":             {"", false, largeFunc, "gt", 2},
		"gte":            {"", false, largeOrEqual, "gte", 2},
		"neqs":           {"", false, notEqSFunc, "neqs", 2},
		"not":            {"", false, notFunc, "not", 1},
		"or":             {"", false, orFunc, "or", -1},
		"and":            {"", false, andFunc, "and", -1},
		"if":             {"", false, ifFunc, "if", -1},
		"len":            {"", false, lenFunc, "len", 1},
		"inn":            {"", false, inFunc, "inn", -1},
		"print":          {"", false, printFunc, "print", -1},
		"add":            {"", false, addFunc, "add", -1},
		"sub":            {"", false, subFunc, "sub", 2},
		"mul":            {"", false, mulFunc, "mul", 2},
		"mod":            {"", false, modFunc, "mod", 2},
		"div":            {"", false, divFunc, "div", 2},
		"pow":            {"", false, powFunc, "pow", 2},
		"neg":            {"", false, negativeFunc, "neg", 1},
		"delete":         {"", false, deleteFunc, "delete", 2},
		"get":            {"", false, getFunc, "get", 2},
		"set":            {"", false, setFunc, "set", 3},
		"set_index":      {"", false, setIndex, "set_index", 3},
		"str_has_prefix": {"", false, hasPrefixFunc, "has_prefix", 2},
		"str_has_suffix": {"", false, hasSuffixFunc, "has_suffix", 2},
		"str_join":       {"", false, joinFunc, "str_join", -1},
		"str_split":      {"", false, splitFunc, "str_split", 3},
		"str_to_upper":   {"", false, toUpperFunc, "str_to_upper", 1},
		"str_to_lower":   {"", false, toLowerFunc, "str_to_lower", 1},
		"str_trim":       {"", false, trimFunc, "str_trim", 1},
		"str_fields":     {"", false, fieldFunc, "str_fields", 1},
		"json_to":        {"", false, jsonEncode, "json_to", 1},
		"to_json_str":    {"", false, jsonEncode, "to_json_str", 1},
		"json_from":      {"", false, jsonDecode, "json_from", 1},
		"to_json_obj":    {"", false, jsonDecode, "to_json_obj", 1},
		"time_now":       {"", false, timeNow, "time_now", 0},
		"time_now_mill":  {"", false, nowTimeMillsec, "time_now_mill", 0},
		"time_from_unix": {"", false, timeFromUnix, "time_from_unix", 1},
		"time_format":    {"", false, timeFormat, "time_format", 2},
		"time_parse":     {"", false, funcTimeParse, "time_parse", 2},
		"type":           {"", false, typeOfFunc, "type", 1},
		"slice_new":      {"", false, newArrFunc, "slice_new", -1},
		"slice_init":     {"", false, sliceInitFunc, "slice_init", -1},
		"slice_cut":      {"", false, arrSliceFunc, "slice_cut", 3},
		"ternary":        {"", false, ternaryFunc, "ternary", 3},
		"string":         {"", false, stringFunc, "string", 1},
		"number":         {"", false, numberFunc, "number", 1},
		"int":            {"", false, intFunc, "int", 1},
		"bool":           {"", false, boolFunc, "bool", 1},
		"bytes":          {"", false, bytesFuncs, "bytes", 1},
		"base64_encode":  {"", false, base64Encode, "base64_encode", 1},
		"base64_decode":  {"", false, base64Decode, "base64_decode", 1},
		"md5_sum":        {"", false, md5SumFunc, "md5", 1},
		"sha256_sum":     {"", false, sha256Func, "sha256", 1},
		"hmac_sha256":    {"", false, hmacSha266Func, "hmac_sha256", 2},
		"hex_encode":     {"", false, hexEncodeFunc, "hex_encode", 1},
		"hex_decode":     {"", false, hexDecodeFunc, "hex_decode", 1},
		"sprintf":        {"", false, sprintfFunc, "sprintf", -1},
		"http_request":   {"", false, httpRequest, "http_request", 5},
		"return":         {"", false, returnFunc, "return", -1},
		"orr":            {"", false, orrFunc, "orr", 2},
		"new":            {"", false, newFunc, "new", 0},
		"all":            {"", false, funcAll, "all", 2},
		"for":            {"", false, funcFor, "for", 2},
		"loop":           {"", false, funcLoop, "loop", -1},
		"go":             {"", false, funcGo, "go", 1},
		"catch":          {"", false, funcCatch, "catch", 1},
		"unwrap":         {"", false, funcUnwrap, "unwrap", 1},
		"boolean":        {"", false, funcBool, "boolean", 1},
		"recover":        {"", false, funcRecover, "recover", 1},
		"sleep":          {"", false, funcSleep, "sleep", 1},
		"range":          {"", false, funcRange, "range", 1},
		"exec":           {"", false, funcExec, "exec", -1},
		"cost":           {"", false, funcCost, "cost", 1},
		"_debug":         {"", false, funcDebug, "_debug", -1},
		"rand":           {"", false, funcRand, "rand", 1},
		"is_empty":       {"", false, funcIsEmpty, "is_empty", 1},
		"printf":         {"", false, funcPrintf, "printf", -1},
		"set_to":         {"", false, funcSetTo, "set_to", 2},
		"benchmark":      {"", false, funcBenchmark, "benchmark", 1},
	}
)

func checkFunction() {
	for s, _ := range funtables {
		if strings.Contains(s, ".") {
			panic("functions must not contain \".\" :" + s)
		}
	}
}
func init() {
	//RegisterFunc("func", defineFunc, 2)

	RegisterFunc("show_doc", funcShowFuncs, 0)
}

var newFunc = FuncDefine(func() any {
	return make(map[string]any)
})

func RegisterDynamicFunc(funName string, argsNum int) {
	if strings.Contains(funName, ".") {
		panic("dynamic function name must not contain '.' :" + funName)
	}
	funtables[funName] = &innerFunc{
		fun: func(ctx *Context, args ...Val) any {
			if ctx.funcs == nil {
				return nil
			}
			f := ctx.funcs[funName]
			if f == nil {
				return nil
			}
			return f(ctx, args...)
		},
		name:    funName,
		argsNum: argsNum,
	}
}

func RegisterFunc(funName string, f ScriptFunc, argsNum int) {
	if strings.Contains(funName, ".") {
		panic("function name must not contain '.':" + funName)
	}
	funtables[funName] = &innerFunc{
		fun:     f,
		name:    funName,
		argsNum: argsNum,
	}
}

type OptFunc func(ctx *Context, args []Val, opt *Options) any

func RegisterFuncWithOpt(funName string, f OptFunc, argsNum int) {
	if strings.Contains(funName, ".") {
		panic("function name must not contain '.':" + funName)
	}
	funtables[funName] = &innerFunc{
		hasOpt: true,
		fun: func(ctx *Context, args ...Val) any {
			var opt *Options
			if len(args) == argsNum+1 {
				opt = args[argsNum].Val(ctx).(*Options)
			}
			return f(ctx, args, opt)
		},
		name:    funName,
		argsNum: argsNum,
	}
}

func RegisterOptFuncDefine2[A any, B any, R any](fname string, f func(ctx *Context, a A, b B, opt *Options) R) {
	RegisterFuncWithOpt(fname, func(ctx *Context, args []Val, opt *Options) any {
		a, _ := args[0].Val(ctx).(A)
		b, _ := args[1].Val(ctx).(B)
		return f(ctx, a, b, opt)
	}, 2)
}

func RegisterOptFuncDefine1[A any, R any](fname string, f func(ctx *Context, a A, opt *Options) R) {
	RegisterFuncWithOpt(fname, func(ctx *Context, args []Val, opt *Options) any {
		a, _ := args[0].Val(ctx).(A)
		return f(ctx, a, opt)
	}, 1)
}

func RegisterOptFuncDefine3[A any, B any, C any, R any](fname string, f func(ctx *Context, a A, b B, c C, opt *Options) R) {
	RegisterFuncWithOpt(fname, func(ctx *Context, args []Val, opt *Options) any {
		a, _ := args[0].Val(ctx).(A)
		b, _ := args[1].Val(ctx).(B)
		c, _ := args[2].Val(ctx).(C)
		return f(ctx, a, b, c, opt)
	}, 3)
}

func RegisterOptFuncDefine4[A any, B any, C any, D any, R any](fname string, f func(ctx *Context, a A, b B, c C, d D, opt *Options) R) {
	RegisterFuncWithOpt(fname, func(ctx *Context, args []Val, opt *Options) any {
		a, _ := args[0].Val(ctx).(A)
		b, _ := args[1].Val(ctx).(B)
		c, _ := args[2].Val(ctx).(C)
		d, _ := args[3].Val(ctx).(D)
		return f(ctx, a, b, c, d, opt)
	}, 4)
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
	ss, ok := arg.([]string)
	if ok {
		return strings.Join(ss, sep)
	}

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
	sb.Grow(length * 3)
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

var orrFunc = FuncDefine2(func(a any, b any) any {
	if a != nil {
		return a
	}
	return b
})

var andFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	for _, arg := range args {
		if !BoolCond(arg.Val(ctx)) {
			return false
		}
	}
	return true
}

var powFunc = FuncDefine2(math.Pow)

var ifFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	var v any
	for _, arg := range args {
		v = arg.Val(ctx)
		if !BoolCond(v) {
			return v
		}
	}
	return v
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
		//if len(args) != 1 {
		//	return nil
		//}
		a, _ := args[0].Val(ctx).(A1)
		return f(a)
	}
}
func FuncDefine1WithCtx[A1 any, R any](f func(ctx *Context, a A1) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		//if len(args) != 1 {
		//	return nil
		//}
		a, _ := args[0].Val(ctx).(A1)
		return f(ctx, a)
	}
}
func FuncDefine[R any](f func() R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		return f()
	}
}
func FuncDefineWithCtx[R any](f func(ctx *Context) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		return f(ctx)
	}
}

func FuncDefine2[A1 any, A2 any, R any](f func(a A1, b A2) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		//if len(args) != 2 {
		//	return nil
		//}
		a, _ := args[0].Val(ctx).(A1)

		b, _ := args[1].Val(ctx).(A2)

		return f(a, b)
	}
}
func FuncDefine2WithCtx[A1 any, A2 any, R any](f func(ctx *Context, a A1, b A2) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		//if len(args) != 2 {
		//	return nil
		//}
		a, _ := args[0].Val(ctx).(A1)

		b, _ := args[1].Val(ctx).(A2)

		return f(ctx, a, b)
	}
}

func FuncDefine3[A1 any, A2 any, A3 any, R any](f func(a A1, b A2, c A3) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		//if len(args) != 3 {
		//	return nil
		//}
		a, _ := args[0].Val(ctx).(A1)

		b, _ := args[1].Val(ctx).(A2)

		c, _ := args[2].Val(ctx).(A3)

		return f(a, b, c)
	}
}
func FuncDefine3WithCtx[A1 any, A2 any, A3 any, R any](f func(ctx *Context, a A1, b A2, c A3) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		//if len(args) != 3 {
		//	return nil
		//}
		a, _ := args[0].Val(ctx).(A1)

		b, _ := args[1].Val(ctx).(A2)

		c, _ := args[2].Val(ctx).(A3)

		return f(ctx, a, b, c)
	}
}
func FuncDefine4[A1 any, A2 any, A3 any, A4 any, R any](f func(a A1, b A2, c A3, d A4) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		//if len(args) != 4 {
		//	return nil
		//}
		a, _ := args[0].Val(ctx).(A1)

		b, _ := args[1].Val(ctx).(A2)

		c, _ := args[2].Val(ctx).(A3)
		d, _ := args[3].Val(ctx).(A4)

		return f(a, b, c, d)
	}
}

func FuncDefine4WithCtx[A1 any, A2 any, A3 any, A4 any, R any](f func(ctx *Context, a A1, b A2, c A3, d A4) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		//if len(args) != 4 {
		//	return nil
		//}
		a, _ := args[0].Val(ctx).(A1)

		b, _ := args[1].Val(ctx).(A2)

		c, _ := args[2].Val(ctx).(A3)
		d, _ := args[3].Val(ctx).(A4)

		return f(ctx, a, b, c, d)
	}
}

func FuncDefine5[A1 any, A2 any, A3 any, A4 any, A5 any, R any](f func(a A1, b A2, c A3, d A4, e A5) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		//if len(args) != 5 {
		//	return nil
		//}
		a, _ := args[0].Val(ctx).(A1)

		b, _ := args[1].Val(ctx).(A2)

		c, _ := args[2].Val(ctx).(A3)
		d, _ := args[3].Val(ctx).(A4)
		e, _ := args[4].Val(ctx).(A5)

		return f(a, b, c, d, e)
	}
}

func FuncDefine5WithCtx[A1 any, A2 any, A3 any, A4 any, A5 any, R any](f func(ctx *Context, a A1, b A2, c A3, d A4, e A5) R) ScriptFunc {
	return func(ctx *Context, args ...Val) any {
		//if len(args) != 5 {
		//	return nil
		//}
		a, _ := args[0].Val(ctx).(A1)

		b, _ := args[1].Val(ctx).(A2)

		c, _ := args[2].Val(ctx).(A3)
		d, _ := args[3].Val(ctx).(A4)
		e, _ := args[4].Val(ctx).(A5)

		return f(ctx, a, b, c, d, e)
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
			arg, _ := arg.Val(ctx).(T)
			argv = append(argv, arg)
		}
		return f(argv...)
	}
}

var setFunc = FuncDefine3(func(m map[string]any, b string, c any) any {
	if m == nil {
		return newErrorf("assign to nil map: k:%v v:%v", b, c)
	}
	m[b] = c
	return nil
})

var setIndex = FuncDefine3(func(m []any, b float64, c any) any {
	idx := int(b)
	if idx >= len(m) {
		return newErrorf("index out of range: k:%v v:%v", b, c)
	}
	m[idx] = c
	return nil
})

var deleteFunc = FuncDefine2(func(m map[string]any, b string) any {
	delete(m, b)
	return nil
})

var getFunc = FuncDefine2(func(m map[string]any, b string) any {
	return m[b]
})

var addFunc = FuncDefineN(func(a ...any) any {
	if len(a) == 0 {
		return nil
	}
	switch v := a[0].(type) {
	case float64:
		sum := v
		for _, va := range a[1:] {
			sum += NumberOf(va)
		}
		return sum
	default:
		sb := strings.Builder{}
		sb.WriteString(StringOf(v))
		for _, va := range a[1:] {
			sb.WriteString(StringOf(va))
		}
		return sb.String()
	}

})

var add2Func = FuncDefine2(func(a, b any) any {
	switch v := a.(type) {
	case float64:
		return v + NumberOf(b)
	case string:
		return v + StringOf(b)
	default:
		return nil
	}
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
var splitFunc = FuncDefine3(func(a, b string, n float64) any {
	vals := strings.SplitAfterN(a, b, int(n))
	va := make([]any, 0, len(vals))
	for _, v := range vals {
		va = append(va, v)
	}
	return va
})

var toUpperFunc = FuncDefine1(strings.ToUpper)

var toLowerFunc = FuncDefine1(strings.ToLower)

var jsonEncode = FuncDefine1(func(a any) any {
	bf := bytes.NewBuffer(nil)
	je := json.NewEncoder(bf)
	je.SetEscapeHTML(false)
	je.Encode(a)
	return unsafe.String(unsafe.SliceData(bf.Bytes()), len(bf.Bytes()))
})

var jsonDecode = FuncDefine1(func(a any) (res any) {
	switch a := a.(type) {
	case string:
		err := json.Unmarshal(unsafe.Slice(unsafe.StringData(a), len(a)), &res)
		if err != nil {
			return newError(err)
		}
		return res
	case []byte:
		err := json.Unmarshal(a, &res)
		if err != nil {
			return newError(err)
		}
		return res
	case nil:
		return nil
	}
	return newErrorf("cannot decode type to json obj %s", reflect.TypeOf(a).String())
})

var timeFormat = FuncDefine2(func(tim time.Time, format string) any {
	return tim.Format(format)
})

var timeNow = FuncDefine(time.Now)

var notFunc = FuncDefine1(func(a any) bool {
	return !BoolCond(a)
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

var largeFunc = FuncDefine2(func(a any, b any) bool {
	return compare(a, b) > 0
})

func compare(a, b any) int {
	switch aa := a.(type) {
	case float64:
		bb := NumberOf(b)
		switch {
		case aa == bb:
			return 0
		case aa < bb:
			return -1
		default:
			return 1
		}
	case int:
		bb := int(NumberOf(b))
		switch {
		case aa == bb:
			return 0
		case aa < bb:
			return -1
		default:
			return 1
		}
	case string:
		bb := StringOf(b)
		return strings.Compare(aa, bb)
	default:
		return 0
	}
}

var largeOrEqual = FuncDefine2(func(a any, b any) bool {
	return compare(a, b) >= 0
})
var lessFunc = FuncDefine2(func(a any, b any) bool {
	return compare(a, b) < 0
})
var lessOrEqual = FuncDefine2(func(a any, b any) bool {
	return compare(a, b) <= 0
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

var newArrFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	size := 0
	cap := 0
	switch len(args) {
	case 0:
	case 1:
		size = int(NumberOf(args[0].Val(ctx)))
	default:
		size = int(NumberOf(args[0].Val(ctx)))
		cap = int(NumberOf(args[1].Val(ctx)))
	}
	if cap < size {
		cap = size
	}
	return make([]any, size, cap)
}

var makeArrFunc = FuncDefine1(func(a float64) any {
	return make([]any, int(a))
})

var sliceInitFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	arr := make([]any, 0, len(args))
	for _, v := range args {
		arr = append(arr, v.Val(ctx))
	}
	return arr
}

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
	ok := BoolCond(args[0].Val(ctx))
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

var intFunc = FuncDefine1(func(a float64) float64 {
	return float64(int(a))
})

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
		return base64EncodeToString(ToBytes(v))
	case []byte:
		return base64EncodeToString(v)
	default:
		return ""
	}
}, "")

var base64Decode = FuncDefine1WithDef(func(a string) any {
	bs, err := base64DecodeString(a)
	if err != nil {
		return newError(err)
	}
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

var hexDecodeFunc = FuncDefine1WithDef(func(a any) any {
	data, err := hex.DecodeString(StringOf(a))
	if err != nil {
		return newError(err)
	}
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
	case nil:
		return 0
	default:
		return float64(lenOfStruct(reflect.ValueOf(a)))
	}
})
var inFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) < 2 {
		return false
	}
	arg := args[0].Val(ctx)
	targets := args[1:]
	for _, target := range targets {
		tv := target.Val(ctx)
		switch tgt := tv.(type) {
		case []any:
			for _, a := range tgt {
				if arg == a {
					return true
				}
			}
		default:
			if arg == tv {
				return true
			}
		}

	}
	return false
}

var returnFunc = FuncDefineN(func(a ...any) any {
	return &Return{
		Var: a,
	}
})

var negativeFunc = FuncDefine1(func(a float64) any {
	return -a
})
var funcAll ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) != 2 {
		return nil
	}
	data := args[0].Val(ctx)
	size := 5
	arr, ok := data.([]any)
	if ok {
		size = len(arr)
	}
	dst := make([]any, 0, size)
	forRangeExec(args[1], ctx, data, func(k, v any, val Val) any {
		data := val.Val(ctx)
		if BoolCond(data) {
			dst = append(dst, v)
		}
		return data
	})
	return dst
}

var funcFor ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) != 2 {
		return nil
	}
	//return lambdaExecMapRange(args[0].Val(ctx).(map[string]any), ctx, args[1].(*lambda), func(k, v any, val Val) any {
	//	return val.Val(ctx)
	//})
	return forRangeExec(args[1], ctx, args[0].Val(ctx), func(k, v any, val Val) any {
		return val.Val(ctx)
	})

}

var defineFunc ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) != 2 {
		return nil
	}
	funName := StringOf(args[0].Val(ctx))
	if !strings.HasPrefix(funName, "$") {
		return newErrorf("func define name must start with '$'")
	}
	lm, ok := args[1].(*lambda)
	ctx.SetFunc(funName, func(ctx *Context, as ...Val) any {
		if ok {
			return lambaCall(lm, ctx, as)
		}
		for i, a := range as {
			ctx.Set("$"+strconv.Itoa(i+1), a.Val(ctx))
		}
		return args[1].Val(ctx)
	})
	return nil
}

func lambaCall(lm *lambda, ctx *Context, as []Val) any {
	argNames := lm.Lefts
	newC := ctx
	if ctx.NewCallEnv {
		newC = ctx.Clone()
	}
	if len(argNames) > len(as) {
		argNames = argNames[:len(as)]
	}

	for i, name := range argNames {
		newC.Set(name, as[i].Val(ctx))
	}
	return lm.Right.Val(newC)
}

var (
	_loopConst = &constraint{
		value: true,
	}
)

var funcLoop ScriptFunc = func(ctx *Context, args ...Val) any {
	var shouldContinue Val
	var doVar Val
	switch len(args) {
	case 1:
		shouldContinue = _loopConst
		doVar = args[0]
	case 2:
		shouldContinue = args[0]
		doVar = args[1]
	default:
		return newErrorf("func loop expects 1 or 2, got %d", len(args))
	}
	for BoolCond(shouldContinue.Val(ctx)) {
		o := doVar.Val(ctx)
		switch o.(type) {
		case *Return, *Error:
			return o
		case *Break:
			return nil
		}
	}
	return nil
}

var funcGo ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) != 1 {
		return nil
	}
	lm, ok := args[0].(*lambda)
	if !ok {
		return newErrorf("go func ,arg should be lambda func")
	}
	goCtx := NewContext(map[string]any{})

	for key, val := range ctx.table {
		goCtx.table[key] = val
	}
	goCtx.funcs = ctx.funcs
	for _, args := range lm.Lefts {
		goCtx.Set(args, ctx.Get(args))
	}
	go lm.Right.Val(goCtx)
	return nil
}

var funcTimeParse = FuncDefine2(func(layout string, val string) any {
	tm, err := time.Parse(layout, val)
	if err != nil {
		return newError(err)
	}
	return tm
})

var funcCatch = FuncDefine1(func(a any) any {

	switch v := a.(type) {
	case *Return, *Error:
		return nil
	case *Result:
		return v.Data
	}
	return a
})

var funcUnwrap = FuncDefine1(func(a any) any {
	res, ok := a.(*Result)
	if ok {
		if res.Err != nil {
			return newError(res.Err)
		}
		return res.Data
	}
	return a
})

var funcBool = FuncDefine1(func(a any) any {
	return BoolOf(a)
})

var funcRecover ScriptFunc = func(ctx *Context, args ...Val) (res any) {
	if len(args) != 1 {
		return nil
	}
	defer func() {
		if r := recover(); r != nil {
			res = r
		}
	}()
	res = args[0].Val(ctx)
	return res
}

var funcSleep = FuncDefine1(func(a float64) any {
	time.Sleep(time.Duration(a) * time.Millisecond)
	return nil
})

var funcRange = FuncDefine1(func(a float64) any {
	return make([]any, int(a))
})

var funcCost ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) != 1 {
		return nil
	}
	start := time.Now()
	args[0].Val(ctx)
	end := time.Now()

	return float64(end.Sub(start).Milliseconds())
}

var funcExec ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) == 0 {
		return nil
	}
	argss := make([]string, 0, len(args)-1)
	for _, arg := range args[1:] {
		argss = append(argss, StringOf(arg.Val(ctx)))
	}
	stdout := new(bytes.Buffer)
	c := StringOf(args[0].Val(ctx))
	vs := strings.Fields(c)
	cmd := exec.Command(vs[0], append(vs[1:], argss...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return newError(err.Error())
	}
	if cmd.ProcessState.ExitCode() != 0 {
		return newError(cmd.ProcessState.ExitCode())
	}
	return stdout.String()
}

var funcDebug ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) == 0 {
		return nil
	}
	val := args[0]
	v := val.Val(ctx)
	argss := make([]any, 0, len(args)-1)
	for _, v := range args[1:] {
		argss = append(argss, v.Val(ctx))
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] %s %s: %v\n", fmt.Sprint(argss...), valTypeOf(val), v)
	return v
}

func valTypeOf(v Val) string {
	switch v := v.(type) {
	case *constraint:
		return fmt.Sprintf("constraint(%v)", v.value)
	case *variable:
		return fmt.Sprintf("variable(%v)", v.varName)
	case *funcVariable:
		return fmt.Sprintf("func(%v)", v.funcName)
	default:
		return "()"
	}
}

var funcIsEmpty = FuncDefine1(func(a any) any {
	switch v := a.(type) {
	case string:
		return v == ""
	case nil:
		return true
	default:
		return false
	}
})

var funcPrintf ScriptFunc = func(ctx *Context, args ...Val) any {
	if len(args) < 1 {
		return nil
	}
	f := StringOf(args[0].Val(ctx))
	av := make([]any, 0, len(args))
	for _, val := range args[1:] {
		av = append(av, val.Val(ctx))
	}
	fmt.Printf(f, av...)
	return nil
}

var (
	_randseed = rand.New(rand.NewSource(time.Now().UnixNano()))
	_randLock = sync.Mutex{}
)

var funcRand = FuncDefine1(func(a float64) []byte {
	_randLock.Lock()
	defer _randLock.Unlock()
	bs := make([]byte, int(a))
	_randseed.Read(bs)
	return bs
})

var funcSetTo ScriptFunc = func(ctx *Context, args ...Val) any {
	v := args[1]
	key := ""
	switch v := v.(type) {
	case *variable:
		key = v.varName
	default:
		key = StringOf(v.Val(ctx))
	}
	vv := args[0].Val(ctx)
	ctx.Set(key, vv)
	return vv
}

var funcBenchmark ScriptFunc = func(ctx *Context, args ...Val) any {
	res := testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			RunLambda(ctx, args[0], nil)
		}
	})
	fmt.Println(res.String(), res.MemString())
	return nil
}

var funcShowFuncs ScriptFunc = func(ctx *Context, args ...Val) any {
	docsObj := []string{}
	//for _, m := range objFuncMap {
	//	for _, o := range m {
	//
	//		docsObj = append(docsObj, fmt.Sprintf("%s::%s\n", o.typeI, o.doc))
	//	}
	//}
	//
	objFuncMap.foreach(func(f *funcMap) bool {
		f.foreach(func(key string, o *objectFunc) bool {
			docsObj = append(docsObj, fmt.Sprintf("%s::%s\n", o.typeI, o.doc))
			return true
		})
		return true
	})

	sort.Strings(docsObj)
	glb := []string{}
	for _, i := range funtables {
		glb = append(glb, fmt.Sprintf("%s()  args: %d\n", i.name, i.argsNum))
	}
	sort.Strings(glb)

	for _, s := range glb {
		fmt.Printf(s)
	}

	for _, s := range docsObj {
		fmt.Printf(s)
	}

	return nil
}
