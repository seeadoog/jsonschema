package expr

import "strings"

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
