package jsonschema

import "strings"

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
}

func funcAppend(ctx Context, args ...Value) interface{} {
	bf := strings.Builder{}
	for _, arg := range args {
		v := arg.Get(ctx)
		bf.WriteString(String(v))
	}
	return bf.String()
}

func funcAdd(ctx Context, args ...Value) interface{} {
	var sum float64 = 0
	for _, arg := range args {
		sum += Number(arg.Get(ctx))
	}
	return sum
}
func funcMul(ctx Context, args ...Value) interface{} {
	var sum float64 = 0
	for _, arg := range args {
		sum *= Number(arg.Get(ctx))
	}
	return sum
}

func funcSub(ctx Context, args ...Value) interface{} {
	if len(args) <= 2 {
		return 0
	}

	return Number(args[0].Get(ctx)) - Number(args[1].Get(ctx))
}

func funcDiv(ctx Context, args ...Value) interface{} {
	if len(args) <= 2 {
		return 0
	}

	return Number(args[0].Get(ctx)) / Number(args[1].Get(ctx))
}

func funcMod(ctx Context, args ...Value) interface{} {
	if len(args) <= 2 {
		return 0
	}

	return int(Number(args[0].Get(ctx))) % int(Number(args[1].Get(ctx)))
}

func funcSplit(ctx Context, args ...Value) interface{} {
	if len(args) < 2 {
		return nil
	}
	str := String(args[0].Get(ctx))
	sep := String(args[1].Get(ctx))
	num := -1
	if len(args) >= 3 {
		num = int(Number(args[2].Get(ctx)))
	}
	return strings.SplitN(str, sep, num)
}

func funcJoin(ctx Context, args ...Value) interface{} {
	if len(args) < 2 {
		return ""
	}
	arri, ok := args[0].Get(ctx).([]string)
	sep := String(args[1].Get(ctx))
	if ok {
		return strings.Join(arri, sep)
	}
	arr, ok := args[0].Get(ctx).([]interface{})
	if !ok {
		return ""
	}
	arrs := make([]string, len(arr))
	for i := range arr {
		arrs[i] = String(arr[i])
	}
	return strings.Join(arrs, sep)
}

func funcTrimPrefix(ctx Context, args ...Value) interface{} {
	if len(args) <= 2 {
		return 0
	}

	return strings.TrimPrefix(String(args[0].Get(ctx)), String(args[1].Get(ctx)))
}

func funcTrimSuffix(ctx Context, args ...Value) interface{} {
	if len(args) <= 2 {
		return 0
	}

	return strings.TrimSuffix(String(args[0].Get(ctx)), String(args[1].Get(ctx)))
}

func funcTrim(ctx Context, args ...Value) interface{} {
	if len(args) <= 2 {
		return 0
	}

	return strings.Trim(String(args[0].Get(ctx)), String(args[1].Get(ctx)))
}
