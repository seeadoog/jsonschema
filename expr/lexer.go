package expr

import (
	"fmt"
	"github.com/seeadoog/jsonschema/v2/expr/ast"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	variables = 0
	constant  = -1
)

type lexer struct {
	tokens []tokenV
	pos    int
	err    []string
	root   ast.Node
}

func (l *lexer) Lex(lval *ast.YySymType) int {
	//TODO implement me
	if l.pos >= len(l.tokens) {
		return 0
	}
	tt := l.tokens[l.pos]
	l.pos++
	switch tt.kind {
	case variables:
		switch tt.tkn {
		case "true":
			lval.SetBool(true)
			return ast.BOOL
		case "false":
			lval.SetBool(false)
			return ast.BOOL
		case "nil":
			return ast.NIL
		}
		nn, err := strconv.ParseFloat(tt.tkn, 64)
		if err == nil {
			lval.SetNum(nn)
			return ast.NUMBER
		}
		lval.SetStr(tt.tkn)
		return ast.IDENT
	case constant:
		lval.SetStr(tt.tkn)
		return ast.STRING
	default:
		return tt.kind
	}
}

func (l *lexer) SetRoot(node ast.Node) {
	l.root = node
}

func (l *lexer) Error(s string) {
	l.err = append(l.err, fmt.Sprintf("%s near: '%v' ", s, l.near()))
}

func (l *lexer) near() string {
	next := l.pos + 5
	pre := l.pos - 5
	if pre < 0 {
		pre = 0
	}
	if next > len(l.tokens) {
		next = len(l.tokens)
	}
	ss := l.tokens[pre:next]
	arr := make([]string, 0, 6)
	for _, s := range ss {
		arr = append(arr, s.tkn)
	}
	return strings.Join(arr, " ")
}

func ParseValueFromNode(node ast.Node, isAccess bool) (Val, error) {
	switch n := node.(type) {
	case *ast.String:
		sp := &strparser{
			str: []rune(n.Val),
		}

		err := sp.parser()
		if err != nil {
			return nil, fmt.Errorf("parse string: %w %s", err, n.Val)
		}
		if len(sp.vals) == 1 && sp.vals[0].kind == 0 {
			return &constraint{
				value: sp.vals[0].val,
			}, nil
		}
		v, err := parseStrVals(sp.vals)
		if err != nil {
			return nil, fmt.Errorf("parse string: %w %s", err, n.Val)
		}
		return v, nil
	case *ast.Number:
		return &constraint{
			value: n.Val,
		}, nil
	case *ast.Bool:
		return &constraint{
			value: n.Val,
		}, nil
	case *ast.Nil:
		return &constraint{}, nil
	case *ast.Variable:
		//var jp *jsonpath.Complied
		//if isJsonPath(n.Name) {
		//	var err error
		//	jp, err = jsonpath.Compile(n.Name)
		//	if err != nil {
		//		return nil, fmt.Errorf("parse vname as jsonpath error:%w", err)
		//	}
		//}
		switch n.Name {
		case "break":
			return &breakVar{}, nil
		}

		return &variable{
			varName: n.Name,
			//varPath: jp,
		}, nil

	case *ast.Access:
		lv, err := ParseValueFromNode(n.L, false)
		if err != nil {
			return nil, fmt.Errorf("binary parse val L error:%w %s", err, lv)
		}
		rv, err := ParseValueFromNode(n.R, true)
		if err != nil {
			return nil, fmt.Errorf("binary parse val R error:%w %s", err, lv)
		}
		rf, ok := rv.(*objFuncVal)
		if ok {
			if allTypeFuncs[rf.funcName] {
				fun := funtables[rf.funcName]
				if fun == nil {
					return nil, fmt.Errorf("binary parse val function is all type funcs but not defined '%s'", rf.funcName)
				}
				return &funcVariable{
					funcName: rf.funcName,
					fun:      fun.fun,
					args:     append([]Val{lv}, rf.args...),
				}, nil
			}
		}
		return &accessVal{
			left:  lv,
			right: rv,
		}, nil
	case *ast.Call:
		if isAccess {
			args := make([]Val, 0, len(n.Args))
			for _, arg := range n.Args {
				argv, err := ParseValueFromNode(arg, false)
				if err != nil {
					return nil, err
				}
				args = append(args, argv)
			}
			return &objFuncVal{
				args:     args,
				funcName: n.Name,
			}, nil
		}
		fun := funtables[n.Name]
		if !strings.HasPrefix(n.Name, "$") {
			if fun == nil {
				return nil, fmt.Errorf("func '%s' is not defined", n.Name)
			}
			if fun.argsNum != -1 && len(n.Args) != fun.argsNum {
				return nil, fmt.Errorf("func '%s' args num should be '%d' but '%d'", n.Name, fun.argsNum, len(n.Args))
			}
		}

		args := make([]Val, 0, len(n.Args))
		for _, arg := range n.Args {
			argv, err := ParseValueFromNode(arg, false)
			if err != nil {
				return nil, err
			}
			args = append(args, argv)
		}
		var f ScriptFunc
		if fun != nil {
			f = fun.fun
		}
		return &funcVariable{
			funcName: n.Name,
			fun:      f,
			args:     args,
		}, nil
	case *ast.Unary:
		val, err := ParseValueFromNode(n.X, false)
		if err != nil {
			return nil, fmt.Errorf("unary parse val error:%w", err)
		}
		switch n.Op {
		case "!":
			return newUnaryValue("!", val, func(ctx *Context, a Val) any {
				return !BoolCond(val.Val(ctx))
			}), nil
		case "-":
			return newUnaryValue("-", val, func(ctx *Context, a Val) any {
				v, _ := val.Val(ctx).(float64)
				return -v
			}), nil
		}
		return nil, fmt.Errorf("unknown unary operator:%s", n.Op)
	case *ast.Binary:
		lv, err := ParseValueFromNode(n.L, false)
		if err != nil {
			return nil, fmt.Errorf("binary parse val L error:%w %s", err, lv)
		}
		rv, err := ParseValueFromNode(n.R, false)
		if err != nil {
			return nil, fmt.Errorf("binary parse val R error:%w %s", err, lv)
		}
		var fun ScriptFunc
		switch n.Op {
		case "+":
			fun = add2Func
		case "-":
			fun = subFunc
		case "*":
			fun = mulFunc
		case "/":
			fun = divFunc
		case "^":
			fun = powFunc
		case "&&":
			fun = andFunc
			return newBinaryValue("&&", lv, rv, func(ctx *Context, a, b Val) any {
				if !BoolCond(a.Val(ctx)) {
					return false
				}
				return BoolCond(b.Val(ctx))
			}), nil
		case "||":
			return newBinaryValue("", lv, rv, func(ctx *Context, a, b Val) any {
				if BoolCond(a.Val(ctx)) {
					return true
				}
				return BoolCond(b.Val(ctx))
			}), nil
		case "==":
			//fun = eqFunc
			return newBinaryValue("==", lv, rv, func(ctx *Context, a, b Val) any {
				return a.Val(ctx) == b.Val(ctx)
			}), nil
		case "<":
			fun = lessFunc
		case "<=":
			fun = lessOrEqual
		case ">":
			fun = largeFunc
		case ">=":
			fun = largeOrEqual
		case "!=":
			return newBinaryValue("!=", lv, rv, func(ctx *Context, a, b Val) any {
				return a.Val(ctx) != b.Val(ctx)
			}), nil
		case "%":
			fun = modFunc

		case "orr":
			return newBinaryValue("orr", lv, rv, func(ctx *Context, a, b Val) any {
				v := a.Val(ctx)
				if v != nil {
					return v
				}
				return b.Val(ctx)
			}), nil
		case ";":
			fun = func(ctx *Context, args ...Val) any {
				var rs any
				for _, arg := range args {
					rs = arg.Val(ctx)
					err := convertToError(rs)
					if err != nil {
						return err
					}
				}
				return rs
			}

		default:
			return nil, fmt.Errorf("unknown operator of binary :%s %s", n.Op, n)
		}
		return &funcVariable{
			funcName: n.Op,
			fun:      fun,
			args:     []Val{lv, rv},
		}, nil
	case *ast.Set:
		//var jp *jsonpath.Complied
		//var err error
		//if isJsonPath(n.L) {
		//	jp, err = jsonpath.Compile(n.L)
		//	if err != nil {
		//		return nil, fmt.Errorf("parse set field error:%w", err)
		//	}
		//}
		key, err := ParseValueFromNode(n.L, false)
		if err != nil {
			return nil, fmt.Errorf("set parse key error:%w %s", err, key)
		}
		val, err := ParseValueFromNode(n.R, false)
		if err != nil {
			return nil, fmt.Errorf("set parse val error:%w", err)
		}
		if n.Const {
			val = tryConvertToConst(val)
			_, ok := val.(*constraint)
			if !ok {
				return nil, fmt.Errorf("set parse val error,val cannot parse as const %T", n.R)
			}
		}
		return &setValue{
			key: key,
			//jp:  jp,
			val: val,
		}, nil
	case *ast.MapSet:
		mapkvs := make([]mapKv, 0, len(n.Kvs))
		for _, kv := range n.Kvs {
			kk, err := ParseValueFromNode(kv.K, false)
			if err != nil {
				return nil, fmt.Errorf("map parse key error:%w", err)
			}
			vv, err := ParseValueFromNode(kv.V, false)
			if err != nil {
				return nil, fmt.Errorf("map parse value error:%w", err)
			}
			mapkvs = append(mapkvs, mapKv{kk, vv})
		}
		mv := &mapDefineVal{
			kvs: mapkvs,
		}
		return mv, nil

	case *ast.ArrDef:
		arrV := &arrDefVal{}
		for i, n2 := range n.V {
			v, err := ParseValueFromNode(n2, false)
			if err != nil {
				return nil, fmt.Errorf("array parse error:%w %v", err, i)
			}
			arrV.vs = append(arrV.vs, v)
		}
		return arrV, nil
	case *ast.ArrAccess:
		arrV := &arrAccessVal{}
		lv, err := ParseValueFromNode(n.L, false)
		if err != nil {
			return nil, fmt.Errorf("array access parse left error:%w %v", err, n.L)
		}
		rv, err := ParseValueFromNode(n.R, false)
		if err != nil {
			return nil, fmt.Errorf("array access parse right error:%w %v", err, n.R)
		}
		arrV.left = lv
		arrV.right = rv
		return arrV, nil

	case *ast.SliceCut:

		v, err := ParseValueFromNode(n.V, false)
		if err != nil {
			return nil, fmt.Errorf("slice cut parse value error:%w %v", err, n.V)
		}
		var st, ed Val
		if n.St != nil {
			st, err = ParseValueFromNode(n.St, false)
			if err != nil {
				return nil, fmt.Errorf("slice cut parse st  error:%w %v", err, n.V)
			}
		}

		if n.Ed != nil {
			ed, err = ParseValueFromNode(n.Ed, false)
			if err != nil {
				return nil, fmt.Errorf("slice cut parse ed error:%w %v", err, n.V)
			}
		}

		return &sliceCutVal{
			st:  st,
			ed:  ed,
			val: v,
		}, nil
	case *ast.Lambda:
		e, err := ParseValueFromNode(n.R, false)
		if err != nil {
			return nil, fmt.Errorf("lambda parse right error:%w %v", err, n.R)
		}
		lm := &lambda{
			Lefts: n.L,
			Right: e,
		}

		//v, err := convertLambda(lm, lm.Right)
		//if err != nil {
		//	return nil, fmt.Errorf("lambda parse right error:%w %v", err, lm.Right)
		//}
		//lm.Right = v
		//return lm, nil

		return lm, nil

	case *ast.Ternary:
		c, err := ParseValueFromNode(n.C, false)
		if err != nil {
			return nil, fmt.Errorf("ternary parse cond error:%w %v", err, n.R)
		}
		l, err := ParseValueFromNode(n.L, false)
		if err != nil {
			return nil, fmt.Errorf("ternary parse left error:%w %v", err, n.R)
		}
		var r Val
		if n.R != nil {
			r, err = ParseValueFromNode(n.R, false)
			if err != nil {
				return nil, fmt.Errorf("ternary parse right error:%w %v", err, n.R)
			}
		}

		return &ternaryVal{
			c: c,
			l: l,
			r: r,
		}, nil

	default:
		return nil, fmt.Errorf("invalid ast.Node type :%T", node)
	}

}

type sliceCutVal struct {
	val Val
	st  Val
	ed  Val
}

func (s *sliceCutVal) Val(c *Context) any {
	//TODO implement me
	f, length := cutterOf(s.val.Val(c))
	if f == nil {
		return nil
	}
	st := 0
	if s.st != nil {
		st = int(NumberOf(s.st.Val(c)))
	}
	ed := length
	if s.ed != nil {
		ed = int(NumberOf(s.ed.Val(c)))
	}
	if st > ed || st < 0 || ed > length {
		return nil
	}
	return f(st, ed)
}

func cutterOf(v any) (func(st, ed int) any, int) {
	switch vs := v.(type) {
	case []any:
		return func(st, ed int) any {
			return vs[st:ed]
		}, len(vs)
	case []byte:
		return func(st, ed int) any {
			return vs[st:ed]
		}, len(vs)
	case string:
		return func(st, ed int) any {
			return vs[st:ed]
		}, len(vs)
	default:
		return nil, 0
	}
}

type mapKv struct {
	k, v Val
}
type mapDefineVal struct {
	kvs []mapKv
}

func (m *mapDefineVal) Val(c *Context) any {
	mm := make(map[string]any)
	for _, kv := range m.kvs {
		key := ""
		vk, ok := kv.k.(*variable)
		if ok {
			//vvv := kv.k.Val(c)
			//_, ok := vvv.(string)
			//if vvv != nil  && {
			//	key = StringOf(vvv)
			//} else {
			key = vk.varName
			//}
		} else {
			key = StringOf(kv.k.Val(c))
		}
		mm[key] = kv.v.Val(c)
	}
	return mm
}

type arrDefVal struct {
	vs []Val
}

func (a *arrDefVal) Val(c *Context) any {
	arr := make([]any, len(a.vs))
	for i, vv := range a.vs {
		arr[i] = vv.Val(c)
	}
	return arr
}

type strval struct {
	kind int
	val  string
}

type stringFmtVal struct {
	vals []Val
}

var arrPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 3)
	},
}

func (s *stringFmtVal) Val(c *Context) any {

	//sb := strings.Builder{}
	//for _, val := range s.vals {
	//	sb.WriteString(StringOf(val.Val(c)))
	//}
	//return sb.String()
	arr := arrPool.Get().([]string)
	//arr := make([]string, 0, len(s.vals))
	for _, val := range s.vals {
		arr = append(arr, StringOf(val.Val(c)))
	}
	l := 0
	for _, s2 := range arr {
		l += len(s2)
	}
	res := make([]byte, 0, l)
	for _, s2 := range arr {
		res = append(res, s2...)
	}
	arrPool.Put(arr[:0])
	return ToString(res)
}

func parseStrVals(vs []*strval) (Val, error) {

	smt := &stringFmtVal{}
	for _, v := range vs {
		switch v.kind {
		case 0:
			smt.vals = append(smt.vals, &constraint{
				value: v.val,
			})
		case 1:
			vv, err := parseValueV(v.val)
			if err != nil {
				return nil, fmt.Errorf("parse fmt value error:%w %s", err, v.val)
			}
			smt.vals = append(smt.vals, vv)
		}
	}
	return smt, nil
}

type strparser struct {
	str   []rune
	pos   int
	vals  []*strval
	token []rune
}

func (s *strparser) next() (rune, bool) {
	if s.pos >= len(s.str) {
		return 0, false
	}
	r := s.str[s.pos]
	s.pos++
	return r, true
}

func (s *strparser) parseVars() error {
	for {
		c, ok := s.next()
		if !ok {
			return fmt.Errorf("unexpected end in string format var ,need '}' to end '${' ")
		}
		switch c {
		case '\'':
			return fmt.Errorf("invalid char ' in string format variable")
		case '}':
			s.appendToken(1)
			return nil
		default:
			s.token = append(s.token, c)
		}
	}
}

func (s *strparser) appendToken(kind int) {
	if len(s.token) == 0 {
		return
	}
	s.vals = append(s.vals, &strval{kind: kind, val: string(s.token)})
	s.token = s.token[:0]

}

func (s *strparser) parser() error {
	for {
		c, ok := s.next()
		if !ok {
			s.appendToken(0)
			return nil
		}
		switch c {
		case '$':
			cc, ok := s.next()
			if !ok {
				s.token = append(s.token, c)
				continue
			}
			if cc == '{' {
				s.appendToken(0)
				err := s.parseVars()
				if err != nil {
					return err
				}
			} else {
				s.token = append(s.token, c)
				s.pos--
			}
		case '\\':
			cc, ok := s.next()
			if !ok {
				return nil
			}
			s.token = append(s.token, cc)

		default:
			s.token = append(s.token, c)
		}
	}
}

type accessVal struct {
	left  Val
	right Val
}

func setForObject(left Val, lv any, right string, c *Context, val any) {
	//rvar, ok := right.(*variable)
	//if !ok {
	//	return
	//}
	//lv := left.Val(c)
	switch parent := lv.(type) {
	case map[string]any:
		parent[right] = val
	case nil:
		pr := map[string]any{}
		set, ok := left.(parentValueSetter)
		if ok {
			set.Set(c, pr)
		}
		pr[right] = val
	case *Result:
		switch right {
		case "err":
			parent.Err = val
		case "data":
			parent.Data = val
		}
	case Setter:
		parent.SetField(c, right, val)
	default:
		setFieldOfStruct(reflect.ValueOf(lv), right, val)

	}
}

func (a *accessVal) Set(c *Context, val any) any {

	switch rv := a.right.(type) {
	case *variable:
		setForObject(a.left, a.left.Val(c), rv.varName, c, val)
		//case *compiledVar:
		//	c.stackSet(rv.index, val)
	}
	return val
}

// abc::b()::c()::d
func (a *accessVal) Val(ctx *Context) any {

	switch v := a.right.(type) {
	case *objFuncVal:
		self := a.left.Val(ctx)
		se, ok := self.(*Error)
		if ok {
			return se
		}
		t := TypeOf(self)
		f := objFuncMap[t]
		if f == nil {
			if ctx.IgnoreFuncNotFoundError {
				return nil
			}
			return &Error{
				Err: fmt.Errorf("type '%v' do not define func '%s'", reflect.TypeOf(self), v.funcName),
			}
		}
		ff := f[v.funcName]
		if ff == nil {
			if ctx.IgnoreFuncNotFoundError {
				return nil
			}
			return &Error{
				Err: fmt.Errorf("type '%v' do not define func '%s'", reflect.TypeOf(self), v.funcName),
			}
		}
		return ff.fun(ctx, self, v.args...)
	case *variable:
		lv := a.left.Val(ctx)
		//lvv, ok := lv.(map[string]any)
		//if ok {
		//	return lvv[v.varName]
		//}
		switch data := lv.(type) {
		case map[string]any:
			return data[v.varName]
		case *Result:
			switch v.varName {
			case "data":
				return data.Data
			case "err":
				return data.Err
			}
			return nil
		case nil:
			return nil
		case Getter:
			return data.GetField(ctx, v.varName)
		default:

			return getFieldOfStruct(reflect.ValueOf(lv), v.varName)
		}
		//return getFieldOfStruct(reflect.ValueOf(lv), v.varName)
	default:
		return nil
	}
}

type Setter interface {
	SetField(ctx *Context, name string, val any)
}

type Getter interface {
	GetField(c *Context, key string) any
}

type IndexGet interface {
	IndexGet(c *Context, key float64) any
}

type IndexSet interface {
	GetIndSet(ctx *Context, key float64, val any)
}

type parentValueSetter interface {
	Set(c *Context, val any) any
}

//// a.b.c
//func (a *accessVal) SetSelf(ctx *Context, v any) {
//	lv := a.left.Val(ctx)
//	if lv == nil {
//		switch lvr := a.left.(type) {
//		case *accessVal:
//			lvrv := lvr.left.Val(ctx)
//			if lvrv == nil {
//				lvr.left.(SetSelf).SetSelf(ctx, map[string]any{})
//			}else{
//				lvr.right.
//				lvrv.(map[string]any)[]
//			}
//		}
//	}
//}

type arrAccessVal struct {
	left  Val
	right Val
}

func (a *arrAccessVal) Set(c *Context, val any) any {
	lv := a.left.Val(c)
	rv := a.right.Val(c)

	switch rvv := rv.(type) {
	case string:
		setForObject(a.left, lv, rvv, c, val)
		return val

	case float64:
		idx := int(rvv)
		parent, ok := lv.([]any)
		if !ok {
			if lv != nil {
				setIndexOfStruct(reflect.ValueOf(lv), idx, val)
				return val
			}
			parent = make([]any, idx+1)
			set, ok := a.left.(parentValueSetter)
			if ok {
				set.Set(c, parent)
			}
		} else {
			if len(parent) <= idx {
				old := parent
				parent = make([]any, idx+1)
				copy(parent, old)
				set, ok := a.left.(parentValueSetter)
				if ok {
					set.Set(c, parent)
				}
			}
		}
		parent[idx] = val
		return val
	case nil:
	}
	return val
}

func (a *arrAccessVal) Val(ctx *Context) any {
	lv := a.left.Val(ctx)
	rv := a.right.Val(ctx)
	switch v := lv.(type) {
	case []any:
		idx := int(NumberOf(rv))

		if idx >= len(v) {
			return nil
		}
		return v[idx]
	case []string:
		idx := int(NumberOf(rv))

		if idx >= len(v) {
			return nil
		}
		return v[idx]
	case map[string]any:
		idx := StringOf(rv)
		return v[idx]
	case nil:
		return nil
	default:
		return getIndexOfSlice(reflect.ValueOf(lv), int(NumberOf(rv)))

	}
	return nil
}

func tryConvertToConst(val Val) Val {
	switch vv := val.(type) {
	case *arrDefVal:
		return tryCovertArrToConst(vv)
	case *mapDefineVal:
		return tryCovertMapToConst(vv)
	}
	return val
}

func tryCovertArrToConst(val *arrDefVal) Val {
	dst := []any{}
	for _, v := range val.vs {
		vv, ok := tryConvertToConst(v).(*constraint)
		if ok {
			dst = append(dst, vv.value)
		} else {
			return val
		}
	}
	return &constraint{
		value: dst,
	}
}
func tryCovertMapToConst(val *mapDefineVal) Val {
	dst := map[string]any{}
	for _, v := range val.kvs {
		//cst, ok := v.(*constraint)
		//if !ok {
		//	return val
		//}
		var ckk any
		ck, ok1 := v.k.(*constraint)
		if ok1 {
			ckk = ck.value
		}
		ck2, ok2 := v.k.(*variable)
		if ok2 {
			ckk = ck2.varName
		}
		if !ok1 && !ok2 {
			return val
		}

		vcv, ok := tryConvertToConst(v.v).(*constraint)
		if ok {
			dst[StringOf(ckk)] = vcv.value
		} else {
			return val
		}
	}
	return &constraint{
		value: dst,
	}
}

type ternaryVal struct {
	c Val
	l Val
	r Val
}

func (t *ternaryVal) Val(c *Context) any {
	if BoolCond(t.c.Val(c)) {
		return t.l.Val(c)
	}
	if t.r == nil {
		return nil
	}

	return t.r.Val(c)
}

type binaryValue struct {
	name string
	fun  func(ctx *Context, a, b Val) any
	l, r Val
}

func (b *binaryValue) Val(c *Context) any {
	return b.fun(c, b.l, b.r)
}

func newBinaryValue(name string, l, r Val, f func(ctx *Context, a, b Val) any) *binaryValue {
	return &binaryValue{
		name: name,
		fun:  f,
		l:    l,
		r:    r,
	}
}

type unaryValue struct {
	name string
	fun  func(ctx *Context, a Val) any
	v    Val
}

func (u *unaryValue) Val(c *Context) any {
	return u.fun(c, u.v)
}

func newUnaryValue(name string, v Val, f func(ctx *Context, a Val) any) *unaryValue {
	return &unaryValue{
		name: name,
		v:    v,
		fun:  f,
	}
}
