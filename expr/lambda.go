package expr

import (
	"fmt"
	"reflect"
)

type lambda struct {
	Lefts []string
	Right Val
}

func (l *lambda) findVarIndex(name string) (int, bool) {
	for i, left := range l.Lefts {
		if left == name {
			return i, true
		}
	}
	return -1, false
}

func (l *lambda) Val(c *Context) any {
	//TODO implement me
	return l
}

func (lv *lambda) setMapKvForLambda(ctx *Context, k, v any) {
	switch len(lv.Lefts) {
	case 0:
	case 1:
		ctx.Set(lv.Lefts[0], v)
	default:
		ctx.Set(lv.Lefts[0], k)
		ctx.Set(lv.Lefts[1], v)
	}
}

var (
	arrKeys = []string{"$key", "$val"}
)
var (
	mapKeys = []string{"$key", "$val"}
)

func forRangeExec(doVal Val, ctx *Context, target any, f func(k, v any, val Val) any) any {
	lm, ok := doVal.(*lambda)
	switch vv := target.(type) {
	case map[string]any:
		if !ok {
			lm = &lambda{
				Lefts: mapKeys,
				Right: doVal,
			}
		}
		return forRangeMapExec(lm, ctx, vv, f)
	case []any:
		if !ok {
			lm = &lambda{
				Lefts: arrKeys,
				Right: doVal,
			}
		}
		return forRangeArr(lm, ctx, vv, f)
	case nil:
		return nil
	default:
		if !ok {
			lm = &lambda{
				Lefts: mapKeys,
				Right: doVal,
			}
		}
		return forRangeStruct(lm, ctx, reflect.ValueOf(vv), f)
	}
	return nil
}

func forRangeMapExec(lv *lambda, ctx *Context, m map[string]any, f func(k, v any, val Val) any) any {
	for k, v := range m {
		ka := any(k)
		lv.setMapKvForLambda(ctx, ka, v)
		vv := f(ka, v, lv.Right)
		if err := convertToError(vv); err != nil {
			return err
		}
		_, ok := vv.(*Break)
		if ok {
			return nil
		}

	}
	return nil
}

func forRangeArr(lv *lambda, ctx *Context, m []any, f func(k, v any, val Val) any) any {
	for k, v := range m {
		ka := any(k)
		lv.setMapKvForLambda(ctx, ka, v)
		vv := f(ka, v, lv.Right)
		if err := convertToError(vv); err != nil {
			return err
		}
		_, ok := vv.(*Break)
		if ok {
			return nil
		}
	}
	return nil
}

func forRangeStruct(lv *lambda, ctx *Context, v reflect.Value, f func(k, v any, val Val) any) any {
	switch v.Kind() {
	case reflect.Map:
		mr := v.MapRange()
		for mr.Next() {
			k := mr.Key().Interface()
			vv := mr.Value().Interface()
			lv.setMapKvForLambda(ctx, structValueToVm(ctx.ForceType, k), structValueToVm(ctx.ForceType, vv))
			vv = f(k, vv, lv.Right)

			if err := convertToError(vv); err != nil {
				return err
			}
			_, ok := vv.(*Break)
			if ok {
				return nil
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			k := float64(i)
			vv := v.Index(i).Interface()
			lv.setMapKvForLambda(ctx, structValueToVm(ctx.ForceType, k), structValueToVm(ctx.ForceType, vv))
			vv = f(k, vv, lv.Right)
			if err := convertToError(vv); err != nil {
				return err
			}
			_, ok := vv.(*Break)
			if ok {
				return nil
			}
		}
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		return forRangeStruct(lv, ctx, v.Elem(), f)
	default:

		if ctx.IgnoreFuncNotFoundError {
			return nil
		}
		return newErrorf("for range at known type %v", v.Type())
	}
	return nil
}

type lambdaStackVal struct {
	stackSize int
	val       Val
}

type lambdaAccessIndex struct {
	i int
}

func (l *lambdaAccessIndex) Val(c *Context) any {
	return c.stack[c.sp-l.i]
}

func convertLambda(lm *lambda, v Val) (Val, error) {
	switch vv := v.(type) {

	case *lambda:

		val, err := convertLambda(vv, lm.Right)
		if err != nil {
			return nil, err
		}
		vv.Right = val
		return val, nil
	case *variable:
		idx, ok := lm.findVarIndex(vv.varName)
		if ok {
			return &lambdaAccessIndex{i: idx}, nil
		}
		return v, nil
	default:
		return nil, fmt.Errorf("bug: lambda contains unknown_type:%T", v)

	}
}

func setLambdaStackVal(lm *lambda, ctx *Context, k any, v any) {
	switch len(lm.Lefts) {
	case 0:
	case 1:
		ctx.stackSet(0, v)
	default:
		ctx.stackSet(1, v)
		ctx.stackSet(0, k)
	}
}

func lambdaExecMapRange(m map[string]any, ctx *Context, lm *lambda, fun func(k, v any, val Val) any) any {
	ctx.sp += len(lm.Lefts)
	for key, val := range m {
		ka := any(key)
		setLambdaStackVal(lm, ctx, ka, val)
		fun(ka, val, lm.Right)
	}
	ctx.sp -= len(lm.Lefts)
	return nil
}
