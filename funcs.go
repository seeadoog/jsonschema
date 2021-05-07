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
	var sum float64 = 0
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
