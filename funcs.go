package jsonschema

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func init() {

	SetFunc("add", funcAdd)
	SetFunc("sub", funcSub)
	SetFunc("mod", funcMod)
	SetFunc("div", funcDiv)
	SetFunc("mul", funcMul)

	SetFunc("str.trimPrefix", funcTrimPrefix)
	SetFunc("str.trimSuffix", funcTrimSuffix)
	SetFunc("str.trim", funcTrim)
	SetFunc("str.split", funcSplit)
	SetFunc("str.join", funcJoin)
	SetFunc("str.replace", funcReplace)
	SetFunc("str.toLower", funcToLower)
	SetFunc("str.toUpper", funcToUpper)
	SetFunc("str.quote", NewFunc11(strconv.Quote))
	SetFunc("append", funcAppend)
	SetFunc("sprintf", funcSprintf)
	SetFunc("or", funcOr)
	SetFunc("delete", funcDelete)

	SetFunc("md5.hex", md5sum)
	SetFunc("sha256.hex", sha256Func)

	SetFunc("map.get", getFunc)
	SetFunc("map.set", funcMapSet)
	SetFunc("map.del", funcDelete)

	SetFunc("time.format", dateFormat)
	SetFunc("time.now", timenow)
	SetFunc("json.to", encodeJSON)
	SetFunc("json.from", decodeJSON)
	SetFunc("new", funcNew)
	SetFunc("tostring", funcToString)
	SetFunc("tonumber", funcToNumber)
	SetFunc("toint", funcToInt)
	SetFunc("tobool", funcToBool)

	SetFunc("rand.new16", funcRand16)

}

func funcAppend(ctx Context, args ...Value) interface{} {
	bf := strings.Builder{}
	for _, arg := range args {
		v := arg.Get(ctx)
		bf.WriteString(StringOf(v))
	}
	return bf.String()
}

func funcAdd(ctx Context, args ...Value) interface{} {
	var sum float64 = 0
	for _, arg := range args {
		sum += NumberOf(arg.Get(ctx))
	}
	return sum
}
func funcMul(ctx Context, args ...Value) interface{} {
	var sum float64 = 1
	for _, arg := range args {
		sum *= NumberOf(arg.Get(ctx))
	}
	return sum
}

func funcSub(ctx Context, args ...Value) interface{} {
	if len(args) <= 2 {
		return 0
	}

	return NumberOf(args[0].Get(ctx)) - NumberOf(args[1].Get(ctx))
}

func funcDiv(ctx Context, args ...Value) interface{} {
	if len(args) <= 2 {
		return 0
	}

	return NumberOf(args[0].Get(ctx)) / NumberOf(args[1].Get(ctx))
}

func funcMod(ctx Context, args ...Value) interface{} {
	if len(args) <= 2 {
		return 0
	}

	return int(NumberOf(args[0].Get(ctx))) % int(NumberOf(args[1].Get(ctx)))
}

func funcSplit(ctx Context, args ...Value) interface{} {
	if len(args) < 2 {
		return nil
	}
	str := StringOf(args[0].Get(ctx))
	sep := StringOf(args[1].Get(ctx))
	num := -1
	if len(args) >= 3 {
		num = int(NumberOf(args[2].Get(ctx)))
	}
	return strings.SplitN(str, sep, num)
}

func funcJoin(ctx Context, args ...Value) interface{} {
	if len(args) < 2 {
		return ""
	}
	arri, ok := args[0].Get(ctx).([]string)
	sep := StringOf(args[1].Get(ctx))
	if ok {
		return strings.Join(arri, sep)
	}
	arr, ok := args[0].Get(ctx).([]interface{})
	if !ok {
		return ""
	}
	arrs := make([]string, len(arr))
	for i := range arr {
		arrs[i] = StringOf(arr[i])
	}
	return strings.Join(arrs, sep)
}

func funcTrimPrefix(ctx Context, args ...Value) interface{} {
	if len(args) < 2 {
		return ""
	}

	return strings.TrimPrefix(StringOf(args[0].Get(ctx)), StringOf(args[1].Get(ctx)))
}

func funcTrimSuffix(ctx Context, args ...Value) interface{} {
	if len(args) < 2 {
		return ""
	}

	return strings.TrimSuffix(StringOf(args[0].Get(ctx)), StringOf(args[1].Get(ctx)))
}

func funcTrim(ctx Context, args ...Value) interface{} {
	if len(args) < 2 {
		return ""
	}

	return strings.Trim(StringOf(args[0].Get(ctx)), StringOf(args[1].Get(ctx)))
}

func funcReplace(ctx Context, args ...Value) interface{} {
	if len(args) < 3 {
		return ""
	}

	return strings.Replace(StringOf(args[0].Get(ctx)), StringOf(args[1].Get(ctx)), StringOf(args[2].Get(ctx)), -1)
}

func funcSprintf(ctx Context, args ...Value) interface{} {
	if len(args) < 1 {
		return nil
	}
	ags := make([]interface{}, 0, len(args)-1)
	for _, value := range args[1:] {
		ags = append(ags, value.Get(ctx))
	}

	return fmt.Sprintf(StringOf(args[0].Get(ctx)), ags...)
}

func funcOr(ctx Context, args ...Value) interface{} {
	for _, arg := range args {
		val := arg.Get(ctx)
		if notNil(val) {
			return val
		}
	}
	return nil
}

func funcDelete(ctx Context, args ...Value) interface{} {
	m, ok := ctx.(map[string]any)
	if !ok {
		return nil
	}
	for _, arg := range args {
		delete(m, StringOf(arg.Get(ctx)))
	}
	return nil
}

func funcToLower(ctx Context, args ...Value) interface{} {
	if len(args) < 1 {
		return ""
	}
	return strings.ToLower(StringOf(args[0].Get(ctx)))
}

func funcToUpper(ctx Context, args ...Value) interface{} {
	if len(args) < 1 {
		return ""
	}
	return strings.ToUpper(StringOf(args[0].Get(ctx)))
}

func md5sum(ctx Context, args ...Value) interface{} {
	sb := bytes.NewBuffer(make([]byte, 0, 20))
	for _, arg := range args {
		s := StringOf(arg.Get(ctx))
		sb.WriteString(s)
	}
	m := md5.Sum(sb.Bytes())
	res := hex.EncodeToString(m[:])
	return res
}

func sha256Func(ctx Context, args ...Value) interface{} {
	sb := bytes.NewBuffer(make([]byte, 0, 20))
	for _, arg := range args {
		s := StringOf(arg.Get(ctx))
		sb.WriteString(s)
	}
	m := sha256.Sum256(sb.Bytes())
	res := hex.EncodeToString(m[:])
	return res
}

func timenow(ctx Context, args ...Value) interface{} {
	return float64(time.Now().Unix())
}

func getFunc(ctx Context, args ...Value) interface{} {
	if len(args) < 2 {
		return nil
	}
	mm, ok := args[0].Get(ctx).(map[string]any)
	if !ok {
		return nil
	}
	k := StringOf(args[1].Get(ctx))
	return mm[k]
}

func NewFunc2[A1, A2 any](f func(a1 A1, a2 A2) any) Func {
	return func(ctx Context, args ...Value) interface{} {
		if len(args) < 2 {
			return nil
		}
		a1, _ := args[0].Get(ctx).(A1)
		//if !ok {
		//	return nil
		//}
		a2, _ := args[1].Get(ctx).(A2)
		//if !ok {
		//	return nil
		//}
		return f(a1, a2)
	}
}

func NewFunc3[A1, A2, A3 any](f func(a1 A1, a2 A2, a3 A3) any) Func {
	return func(ctx Context, args ...Value) interface{} {
		if len(args) < 3 {
			return nil
		}
		a1, _ := args[0].Get(ctx).(A1)
		//if !ok {
		//	return nil
		//}
		a2, _ := args[1].Get(ctx).(A2)
		//if !ok {
		//	return nil
		//}

		a3, _ := args[2].Get(ctx).(A3)
		//if !ok {
		//	return nil
		//}
		return f(a1, a2, a3)
	}
}

func NewFunc1[A1 any](f func(a1 A1) any) Func {
	return func(ctx Context, args ...Value) interface{} {
		if len(args) < 1 {
			return nil
		}
		a1, _ := args[0].Get(ctx).(A1)
		//if !ok {
		//	return nil
		//}

		return f(a1)
	}
}

func NewFunc11[A1 any, R1 any](f func(a1 A1) R1) Func {
	return func(ctx Context, args ...Value) interface{} {
		if len(args) < 1 {
			return nil
		}
		a1, _ := args[0].Get(ctx).(A1)
		//if !ok {
		//	return nil
		//}

		return f(a1)
	}
}

var hmacSha256 Func = NewFunc2(func(v any, secret string) any {
	h := hmac.New(sha256.New, []byte(secret))
	switch v := v.(type) {
	case string:
		h.Write([]byte(v))
	case []byte:
		h.Write(v)
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
})

var dateFormat = NewFunc2(func(a1 any, a2 string) any {
	switch a := a1.(type) {
	case float64:
		return time.Unix(int64(a), 0).Format(a2)
	case time.Time:
		return a.Format(a2)
	}
	return nil
})

var decodeJSON = NewFunc1(func(a1 any) (res any) {
	switch a1 := a1.(type) {
	case []byte:
		err := json.Unmarshal(a1, &res)
		if err != nil {
			return nil
		}
	case string:
		err := json.Unmarshal([]byte(a1), &res)
		if err != nil {
			return nil
		}
	}
	return res
})

var encodeJSON = NewFunc1(func(a1 any) any {
	data, _ := json.Marshal(a1)
	return string(data)
})

var funcNew Func = func(ctx Context, args ...Value) interface{} {
	return make(map[string]any)
}

var funcToNumber = NewFunc1(func(a1 any) any {
	return NumberOf(a1)
})

var funcToInt = NewFunc1(func(a1 any) any {
	return float64(int(NumberOf(a1)))
})

var funcToBool = NewFunc1(func(a1 any) any {
	return BoolOf(a1)
})

var funcRand16 Func = func(ctx Context, args ...Value) interface{} {
	bs := make([]byte, 16)
	rand.Read(bs)
	return hex.EncodeToString(bs)
}

var funcMapSet = NewFunc3(func(a1 map[string]any, a2 any, a3 any) any {
	if a1 == nil {
		return nil
	}
	a1[StringOf(a2)] = a3
	return nil
})

var funcToString = NewFunc1(func(a1 any) any {
	return StringOf(a1)
})
