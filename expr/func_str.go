package expr

import (
	"strings"
	"time"
)

var newStringBuilder ScriptFunc = func(ctx *Context, args ...Val) any {
	cap := 32
	if len(args) > 0 {
		c := int(NumberOf(args[0].Val(ctx)))
		if c > 0 {
			cap = c
		}
	}
	sb := &strings.Builder{}
	sb.Grow(cap)
	return sb
}

var nowTimeMillsec ScriptFunc = func(ctx *Context, args ...Val) any {
	return float64(time.Now().Nanosecond() / 1e6)
}

var timeFromUnix = FuncDefine1(func(a float64) time.Time {
	return time.Unix(int64(a), 0)
})
