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
	"strings"
	"time"
)

func init() {
	SetFunc("split", funcSplit)
	SetFunc("join", funcJoin)
	SetFunc("add", funcAdd)
	SetFunc("append", funcAppend)
	SetFunc("sub", funcSub)
	SetFunc("mod", funcMod)
	SetFunc("div", funcDiv)
	SetFunc("mul", funcMul)
	SetFunc("trimPrefix", funcTrimPrefix)
	SetFunc("trimSuffix", funcTrimSuffix)
	SetFunc("trim", funcTrim)
	SetFunc("replace", funcReplace)
	SetFunc("sprintf", funcSprintf)
	SetFunc("or", funcOr)
	SetFunc("delete", funcDelete)
	SetFunc("toLower", funcToLower)
	SetFunc("toUpper", funcToUpper)
	SetFunc("md5sum", md5sum)
	SetFunc("nowtime", timenow)
	SetFunc("get", getFunc)
	SetFunc("dateFormat", dateFormat)
	SetFunc("toJson", encodeJSON)
	SetFunc("fromJson", decodeJSON)
	SetFunc("new", funcNew)

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
	for _, arg := range args {
		delete(ctx, StringOf(arg.Get(ctx)))
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

func newFunc2[A1, A2 any](f func(a1 A1, a2 A2) any) Func {
	return func(ctx Context, args ...Value) interface{} {
		if len(args) < 2 {
			return nil
		}
		a1, ok := args[0].Get(ctx).(A1)
		if !ok {
			return nil
		}
		a2, ok := args[1].Get(ctx).(A2)
		if !ok {
			return nil
		}
		return f(a1, a2)
	}
}

func newFunc1[A1 any](f func(a1 A1) any) Func {
	return func(ctx Context, args ...Value) interface{} {
		if len(args) < 1 {
			return nil
		}
		a1, ok := args[0].Get(ctx).(A1)
		if !ok {
			return nil
		}

		return f(a1)
	}
}

var hmacSha256 Func = newFunc2(func(v any, secret string) any {
	h := hmac.New(sha256.New, []byte(secret))
	switch v := v.(type) {
	case string:
		h.Write([]byte(v))
	case []byte:
		h.Write(v)
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
})

var dateFormat = newFunc2(func(a1 any, a2 string) any {
	switch a := a1.(type) {
	case float64:
		return time.Unix(int64(a), 0).Format(a2)
	case time.Time:
		return a.Format(a2)
	}
	return nil
})

var decodeJSON = newFunc1(func(a1 any) (res any) {
	switch a1 := a1.(type) {
	case []byte:
		json.Unmarshal(a1, &res)
	case string:
		json.Unmarshal([]byte(a1), &res)
	}
	return nil
})

var encodeJSON = newFunc1(func(a1 any) any {
	data, _ := json.Marshal(a1)
	return string(data)
})

var funcNew Func = func(ctx Context, args ...Value) interface{} {
	return make(map[string]any)
}
