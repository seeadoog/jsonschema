package expr

import (
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

func (l *lambda) Set(c *Context, v any) {
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

	//lmc, ok := doVal.(*compiledLambda)
	//if ok {
	//	return forRangeCompiledLambda(lmc, ctx, target, f)
	//}

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

func RunLambda(ctx *Context, v Val, args []any) any {
	lm, ok := v.(*lambda)
	if !ok {
		return v.Val(ctx)
	}
	lefts := lm.Lefts
	if len(lefts) > len(args) {
		lefts = lefts[:len(args)]
	}
	for i, left := range lm.Lefts {
		ctx.Set(left, args[i])
	}
	return lm.Right.Val(ctx)
}

//func (c *Context) stackSet(i int, val any) {
//	i = c.sp - i
//	if len(c.stack) <= i {
//		old := c.stack
//		c.stack = make([]any, (i+1)*2)
//		copy(c.stack, old)
//	}
//	c.stack[i] = val
//}

//func (c *Context) InitStackSize(s int) {
//	c.stack = make([]any, s)
//}

//func (c *Context) SetStackAndEnv(cc *StackRootValue, name string, v any) {
//	n, ok := cc.ctx.getIndexOnly(name)
//	if ok {
//		if c.sp == 0 {
//			c.stackSet(n-cc.ctx.sp, v)
//
//		} else {
//			c.stackSet(n, v)
//
//		}
//	}
//	c.Set(name, v)
//}
//
//func (c *Context) stackGet(i int) any {
//	i = c.sp - i
//	if i >= len(c.stack) || i < 0 {
//		return nil
//	}
//	return c.stack[i]
//}
//func (c *Context) GetFromStack(v Val, key string) any {
//	srv, ok := v.(*StackRootValue)
//	if !ok {
//		return nil
//	}
//	return c.stackGet(srv.ctx.getIndex(key))
//}

//
//func forRangeCompiledLambda(lm *compiledLambda, ctx *Context, val any, operator func(k, v any, val Val) any) any {
//	ctx.sp += lm.ctx.sp
//	defer func() {
//		ctx.sp -= lm.ctx.sp
//	}()
//	switch vv := val.(type) {
//	case map[string]any:
//		return forRangeMapCompiledLambda(lm, ctx, vv, operator)
//	case []any:
//		return forRangeArrayCompiledLambda(lm, ctx, vv, operator)
//	}
//	return newErrorf("for range at known type %v", reflect.TypeOf(val))
//}
//
//func forRangeMapCompiledLambda(lm *compiledLambda, ctx *Context, val map[string]any, operator func(k, v any, val Val) any) any {
//	for key, v := range val {
//		ka := any(key)
//		lm.SetMapKeyV(ctx, ka, v)
//		res := operator(ka, v, lm.Right)
//		if err := convertToError(res); err != nil {
//			return err
//		}
//		_, ok := res.(*Break)
//		if ok {
//			return nil
//		}
//	}
//	return nil
//}
//
//func forRangeArrayCompiledLambda(lm *compiledLambda, ctx *Context, val []any, operator func(k, v any, val Val) any) any {
//	for key, v := range val {
//		ka := any(key)
//		lm.SetMapKeyV(ctx, ka, v)
//		res := operator(ka, v, lm.Right)
//		if err := convertToError(res); err != nil {
//			return err
//		}
//		_, ok := res.(*Break)
//		if ok {
//			return nil
//		}
//	}
//	return nil
//}
//
//func compiledLambdaCall(lm *compiledLambda, ctx *Context, args []Val) any {
//	ctx.sp += lm.ctx.sp
//	defer func() {
//		ctx.sp -= lm.ctx.sp
//	}()
//	for i, arg := range args {
//		if i < len(lm.Lefts) {
//			ctx.stackSet(lm.Lefts[i], arg.Val(ctx))
//		}
//	}
//	return lm.Right.Val(ctx)
//}
//
//type compileContext struct {
//	parent    *compileContext
//	nameIndex map[string]int
//	sp        int
//	n2        bool
//}
//
//func (c *compileContext) getIndexOnlyCtx(name string) (*compileContext, int, bool) {
//	i, ok := c.nameIndex[name]
//	if ok {
//		return c, i, true
//	}
//	if c.parent != nil {
//		pc, i, ok := c.parent.getIndexOnlyCtx(name)
//		if ok {
//			return pc, i, true
//		}
//
//	}
//	return nil, i, false
//}
//
//func (c *compileContext) getIndexOnly(name string) (int, bool) {
//	i, ok := c.nameIndex[name]
//	if ok {
//		return i, true
//	}
//	//if c.parent != nil {
//	//	i, ok = c.parent.getIndexOnly(name)
//	//	if ok {
//	//		return i, true
//	//	}
//	//
//	//}
//	return i, false
//}
//
//func (c *compileContext) getIndex(name string) int {
//	i, ok := c.nameIndex[name]
//	if ok {
//		return i
//	}
//	i = c.sp
//	c.nameIndex[name] = i
//	c.sp++
//	return i
//}
//
//// a = 5 ; b = 6 ; c = 9 ; a = a => a
//
//type compiledVar struct {
//	name  string
//	index int
//	pctx  *compileContext
//	ctx   *compileContext
//}
//
//func (c *compiledVar) Set(ctx *Context, val any) any {
//	//TODO implement me
//	ctx.stackSet(c.getIndex(), val)
//	return val
//}
//func (c *compiledVar) getIndex() int {
//	//if c.pctx != nil {
//	//	offset := 0
//	//	pc := c.ctx
//	//	for pc != nil && pc != c.pctx {
//	//		offset += pc.sp
//	//		pc = pc.parent
//	//	}
//	//	return offset + c.index
//	//}
//	//return c.index
//	return getCompiledVarIndex(c.ctx, c.pctx, c.index)
//}
//
//func getCompiledVarIndex(ctx, pctx *compileContext, idx int) int {
//	if pctx != nil {
//		offset := 0
//		pc := ctx
//		for pc != nil && pc != pctx {
//			offset += pc.sp
//			pc = pc.parent
//		}
//		return offset + idx
//	}
//	return idx
//}
//
//func (c *compiledVar) Val(ctx *Context) any {
//	index := c.getIndex()
//
//	v := ctx.stackGet(index)
//	if v == nil {
//		v = ctx.Get(c.name)
//		ctx.stackSet(index, v)
//	}
//	return v
//}
//
//type StackRootValue struct {
//	ctx *compileContext
//	val Val
//}
//
//func (s *StackRootValue) Val(c *Context) any {
//	c.sp = s.ctx.sp
//	return s.val.Val(c)
//}
//
//func (s *StackRootValue) StackSize() int {
//	return s.ctx.sp
//}
//
//func newCompileContext(parent *compileContext) *compileContext {
//	return &compileContext{
//		parent:    parent,
//		nameIndex: make(map[string]int),
//	}
//}
//
//func compileRootValue(val Val) (res *StackRootValue, err error) {
//	ctx := newCompileContext(nil)
//	v, err := compileValue(ctx, val)
//	if err != nil {
//		return nil, err
//	}
//	return &StackRootValue{ctx: ctx, val: v}, nil
//}
//
//type compiledLambda struct {
//	ctx   *compileContext
//	Left  []string
//	Lefts []int
//	Right Val
//}
//
//func (c *compiledLambda) SetMapKeyV(ctx *Context, k any, v any) {
//	switch len(c.Lefts) {
//	case 0:
//	case 1:
//		//idx, ok := c.ctx.getIndexOnly(c.Left[0])
//		//if !ok {
//		//	panic("get index failed")
//		//}
//		idx := c.Lefts[1]
//		ctx.stackSet(idx, v)
//	case 2:
//
//		ctx.stackSet(c.Lefts[0], k)
//		ctx.stackSet(c.Lefts[1], v)
//	}
//}
//
//func (c2 *compiledLambda) Val(c *Context) any {
//	//TODO implement me
//	return c2
//}
//
//type compiledFuncVariable struct {
//	funcName string
//	index    int
//	fun      func(ctx *Context, args ...Val) any
//	args     []Val
//	ctx      *compileContext
//	pctx     *compileContext
//}
//
//// $fb = {d} => a ; $fb()
//func (c *compiledFuncVariable) Val(ctx *Context) any {
//	if c.fun == nil {
//		lm, ok := ctx.stackGet(getCompiledVarIndex(c.ctx, c.pctx, c.index)).(*compiledLambda)
//		if ok {
//			return compiledLambdaCall(lm, ctx, c.args)
//		}
//		if ctx.funcs != nil {
//			f := ctx.funcs[c.funcName]
//			if f != nil {
//				return f(ctx, c.args...)
//			}
//		}
//		if ctx.IgnoreFuncNotFoundError {
//			return nil
//		}
//		return newErrorf("function '%s' not found in table", c.funcName)
//	}
//	return c.fun(ctx, c.args...)
//}
//
//func compileValue(ctx *compileContext, val Val) (res Val, err error) {
//	switch v := val.(type) {
//	case *lambda:
//		newc := newCompileContext(ctx)
//		lfs := make([]int, len(v.Lefts))
//		for i, left := range v.Lefts {
//			lfs[i] = newc.getIndex(left)
//		}
//		newv, err := compileValue(newc, v.Right)
//		if err != nil {
//			return nil, err
//		}
//
//		return &compiledLambda{
//			ctx:   newc,
//			Left:  v.Lefts,
//			Right: newv,
//			Lefts: lfs,
//		}, nil
//	case *variable:
//		n, ok := ctx.getIndexOnly(v.varName)
//		if ok {
//			return &compiledVar{
//				name:  v.varName,
//				index: n,
//			}, nil
//		}
//		if ctx.parent != nil {
//			pc, idx, ok := ctx.parent.getIndexOnlyCtx(v.varName)
//			if ok {
//				return &compiledVar{
//					name:  v.varName,
//					index: idx,
//					pctx:  pc,
//					ctx:   ctx,
//				}, nil
//			}
//		}
//		idx := ctx.getIndex(v.varName)
//		return &compiledVar{
//			name:  v.varName,
//			index: idx,
//		}, nil
//	case *setValue:
//		v.key, err = compileValue(ctx, v.key)
//		if err != nil {
//			return nil, err
//		}
//		v.val, err = compileValue(ctx, v.val)
//		if err != nil {
//			return nil, err
//		}
//
//		return v, nil
//	case *accessVal:
//		//cr, err := compileValue(ctx, v.right)
//		//if err != nil {
//		//	return nil, err
//		//}
//		v.left, err = compileValue(ctx, v.left)
//		if err != nil {
//			return nil, err
//		}
//		rf, ok := v.right.(*objFuncVal)
//		if ok {
//			v.right, err = compileValue(ctx, rf)
//			if err != nil {
//				return nil, err
//			}
//		}
//		//v.right = cr
//		return v, nil
//	case *arrAccessVal:
//		v.left, err = compileValue(ctx, v.left)
//		if err != nil {
//			return nil, err
//		}
//		return v, nil
//	case *constraint:
//		return v, nil
//
//	case *funcVariable:
//		cf := &compiledFuncVariable{
//			args: make([]Val, len(v.args)),
//		}
//		cf.funcName = v.funcName
//		cf.fun = v.fun
//
//		n, ok := ctx.getIndexOnly(v.funcName)
//		if ok {
//			cf.index = n
//		} else {
//			if ctx.parent != nil {
//				pc, idx, ok := ctx.parent.getIndexOnlyCtx(v.funcName)
//				if ok {
//					cf.pctx = pc
//					cf.ctx = ctx
//					cf.index = idx
//				} else {
//					cf.index = ctx.getIndex(v.funcName)
//				}
//			} else {
//				cf.index = ctx.getIndex(v.funcName)
//			}
//		}
//
//		for i, arg := range v.args {
//			cv, err := compileValue(ctx, arg)
//			if err != nil {
//				return nil, err
//			}
//			cf.args[i] = cv
//		}
//		return cf, nil
//
//	case *mapDefineVal:
//		for i, kv := range v.kvs {
//			nv, err := compileValue(ctx, kv.v)
//			if err != nil {
//				return nil, err
//			}
//			v.kvs[i].v = nv
//		}
//		return v, nil
//	case *arrDefVal:
//		for i, vv := range v.vs {
//			v.vs[i], err = compileValue(ctx, vv)
//			if err != nil {
//				return nil, err
//			}
//		}
//		return v, nil
//	case *objFuncVal:
//		for i, arg := range v.args {
//			v.args[i], err = compileValue(ctx, arg)
//			if err != nil {
//				return nil, err
//			}
//		}
//		return v, nil
//	case *unaryValue:
//		v.v, err = compileValue(ctx, v.v)
//		return v, err
//	case *binaryValue:
//		v.l, err = compileValue(ctx, v.l)
//		if err != nil {
//			return nil, err
//		}
//		v.r, err = compileValue(ctx, v.r)
//		return v, err
//	case *sliceCutVal:
//		v.val, err = compileValue(ctx, v.val)
//		return v, err
//	case *stringFmtVal:
//		for i, vv := range v.vals {
//			v.vals[i], err = compileValue(ctx, vv)
//			if err != nil {
//				return nil, err
//			}
//		}
//		return v, nil
//	case *breakVar:
//		return v, nil
//	case *ternaryVal:
//		v.c, err = compileValue(ctx, v.c)
//		if err != nil {
//			return nil, err
//		}
//		v.l, err = compileValue(ctx, v.l)
//		if err != nil {
//			return nil, err
//		}
//		v.r, err = compileValue(ctx, v.r)
//		if err != nil {
//			return nil, err
//		}
//
//		return v, err
//
//	default:
//		panic("unknown val type:" + reflect.TypeOf(val).String())
//	}
//}
