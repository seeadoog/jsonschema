package expr

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/seeadoog/jsonschema/v2/jsonpath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

/*
"if":"eq(name,5)"

*/

type Context struct {
	table map[string]any
	funcs map[string]ScriptFunc
}

func NewContext(table map[string]any) *Context {
	if table == nil {
		table = make(map[string]any)
	}
	return &Context{
		table: table,
	}
}

func (c *Context) GetJP(jp *jsonpath.Complied) interface{} {
	res, ok := jp.Get(c.table)
	if !ok {
		return nil
	}
	return res
}
func (c *Context) Get(key string) interface{} {
	return c.table[key]
}
func (c *Context) Set(key string, value interface{}) {
	c.table[key] = value
}

func (c *Context) SetJP(key *jsonpath.Complied, value interface{}) error {
	return key.Set2(c.table, value)
}

func (c *Context) Delete(key string) {
	delete(c.table, key)
}

func (c *Context) SetFunc(key string, fn ScriptFunc) {
	if funtables[key] == nil {
		panic(fmt.Sprintf("func '%s' not registerd by RegisterDynamicFunc", key))
	}
	if c.funcs == nil {
		c.funcs = make(map[string]ScriptFunc)
	}
	c.funcs[key] = fn
}

func (c *Context) Exec(e Expr) error {
	err := e.Exec(c)
	if err != nil {
		if err == errBreak || err == errReturn {
			return nil
		}
	}
	return err
}

type Expr = exp
type exp interface {
	Exec(c *Context) error
}

type Val interface {
	Val(c *Context) any
}

type variable struct {
	varName string
	varPath *jsonpath.Complied
}

func (v *variable) Val(c *Context) any {
	return c.GetJP(v.varPath)
}

type constraint struct {
	value any
}

func (c *constraint) Val(ctx *Context) any {
	return c.value
}

type ScriptFunc func(ctx *Context, args ...Val) any
type funcVariable struct {
	fun  func(ctx *Context, args ...Val) any
	args []Val
}

func (c *funcVariable) Val(ctx *Context) any {
	return c.fun(ctx, c.args...)
}

type ifCond struct {
	cond     Val
	thenCond exp
	elseCond exp
}

func (i *ifCond) Exec(c *Context) error {
	if BoolOf(i.cond.Val(c)) {
		if i.thenCond != nil {
			return i.thenCond.Exec(c)
		}
	} else {
		if i.elseCond != nil {
			return i.elseCond.Exec(c)
		}
	}
	return nil
}

type setCond struct {
	varName   string
	nameJPath *jsonpath.Complied
	val       Val
}

func (s *setCond) Exec(c *Context) error {
	return c.SetJP(s.nameJPath, s.val.Val(c))
}

type callCond struct {
	fun *funcVariable
}

func (c *callCond) Exec(ctx *Context) error {
	c.fun.Val(ctx)
	return nil
}

type forRange struct {
	target  Val
	keyName string
	valName string
	do      exp
}

var (
	errBreak  = errors.New("break")
	errReturn = errors.New("return")
)

func (f *forRange) Exec(c *Context) error {
	v := f.target.Val(c)
	length := 0
	var valueOf func(i int) any
	switch v := v.(type) {
	case []any:
		length = len(v)
		valueOf = func(i int) any {
			return v[i]
		}
	case []string:
		length = len(v)
		valueOf = func(i int) any {
			return v[i]
		}
	case []float64:
		length = len(v)
		valueOf = func(i int) any {
			return v[i]
		}
	case map[string]interface{}:
		for i, a := range v {
			c.Set(f.keyName, i)
			c.Set(f.valName, a)
			err := f.do.Exec(c)
			if err == errBreak {
				return nil
			}
			if err != nil {
				return err
			}
		}
		return nil
	}
	if valueOf != nil {
		for i := 0; i < length; i++ {
			c.Set(f.keyName, i)
			c.Set(f.valName, valueOf(i))
			err := f.do.Exec(c)
			if err == errBreak {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type errExpr struct {
	err error
}

func (b *errExpr) Exec(c *Context) error {
	return b.err
}

type switchExpr struct {
	val   Val
	cases map[any]exp
	def   exp
}

func (s *switchExpr) Exec(c *Context) error {
	v := s.val.Val(c)
	expr := s.cases[v]
	if expr == nil {
		if s.def != nil {
			return s.def.Exec(c)
		}
		return nil
	}
	return expr.Exec(c)
}

var (
	setCondReg = regexp.MustCompile(`([$_\-0-9a-zA-Z.\[\]]+\s*)=\s*(.*)`)

	funcCallReg = regexp.MustCompile(`^[$._0-9a-zA-Z]+\s*\((.*)\)`)
)

func isSetCond(e string) bool {
	return setCondReg.MatchString(e)
}
func isFuncCall(e string) bool {
	return funcCallReg.MatchString(e)
}

type exps struct {
	exps []exp
}

func (e *exps) Exec(c *Context) error {
	for _, e2 := range e.exps {
		err := e2.Exec(c)
		if err != nil {
			return err
		}
	}
	return nil
}

type ExpParseFunc func(o map[string]any, val any) (Expr, error)

var (
	expParserFactory = map[string]ExpParseFunc{}
)

func init() {
	expParserFactory["if"] = parseIf
	expParserFactory["for"] = parseForRange
	expParserFactory["switch"] = parseSwitch

}

func RegisterExp(name string, exp ExpParseFunc) {
	expParserFactory[name] = exp
}

var parseIf ExpParseFunc = func(o map[string]any, val any) (exp, error) {
	s, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("if val must be a string,but %v", val)
	}
	v, err := parseValueV(s)
	if err != nil {
		return nil, fmt.Errorf("parse if value error:%w %v", err, val)
	}

	iff := &ifCond{
		cond:     v,
		thenCond: nil,
		elseCond: nil,
	}
	thenRaw := o["then"]
	if thenRaw != nil {
		exp, err := ParseFromJSONObj(thenRaw)
		if err != nil {
			return nil, fmt.Errorf("parse if then error:%w %v", err, thenRaw)
		}
		iff.thenCond = exp
	}
	elseRaw := o["else"]
	if elseRaw != nil {
		exp, err := ParseFromJSONObj(elseRaw)
		if err != nil {
			return nil, fmt.Errorf("parse if else error:%w %v", err, elseRaw)
		}
		iff.elseCond = exp
	}

	return iff, nil
}

var (
	forRegexp = regexp.MustCompile(`^(\w+)\s*,\s*(\w+)\s*in\s*(.+)$`)
)

var parseForRange ExpParseFunc = func(o map[string]any, val any) (exp, error) {
	s, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("for val must be a string,but %v", val)
	}

	if !forRegexp.MatchString(s) {
		return nil, fmt.Errorf("invalid for exp %v", s)
	}

	values := forRegexp.FindAllStringSubmatch(s, -1)
	if len(values) == 0 || len(values[0]) < 3 {
		return nil, fmt.Errorf("invalid for exp reg err %v", s)
	}
	vv, err := parseValueV(values[0][3])
	if err != nil {
		return nil, fmt.Errorf("parse for value as val error:%w %v", err, s)
	}

	do, err := ParseFromJSONObj(o["do"])
	if err != nil {
		return nil, fmt.Errorf("parse for::do exp error:%w %v", err, o)
	}
	e := &forRange{
		target:  vv,
		keyName: values[0][1],
		valName: values[0][2],
		do:      do,
	}
	return e, nil
}

type casesExprs struct {
	val  Val
	expr exp
}

type swtichCasesExpr struct {
	cases       []casesExprs
	defaultExpr Expr
}

func (s *swtichCasesExpr) Exec(c *Context) error {
	for _, exprs := range s.cases {
		if BoolOf(exprs.val.Val(c)) {
			return exprs.expr.Exec(c)
		}
	}
	if s.defaultExpr != nil {
		return s.defaultExpr.Exec(c)
	}
	return nil
}

func parseSwitchExpr(o map[string]any, val map[string]any) (exp, error) {
	switchCases := &swtichCasesExpr{}
	for cases, exp := range val {
		val, err := parseValueV(cases)
		if err != nil {
			return nil, fmt.Errorf("parse switchcases cases error:%w %v", err, cases)
		}
		expr, err := ParseFromJSONObj(exp)
		if err != nil {
			return nil, fmt.Errorf("parse switchcases expr error:%w %v:%v", err, cases, exp)
		}
		switchCases.cases = append(switchCases.cases, casesExprs{
			val:  val,
			expr: expr,
		})
	}

	def := o["default"]
	if def != nil {
		defExpr, err := ParseFromJSONObj(def)
		if err != nil {
			return nil, fmt.Errorf("parse switchcases default expr error:%w %v", err, def)
		}
		switchCases.defaultExpr = defExpr
	}
	return switchCases, nil
}

var parseSwitch ExpParseFunc = func(o map[string]any, val any) (exp, error) {
	str, ok := val.(string)
	if !ok {
		ov, ok := val.(map[string]any)
		if ok {
			return parseSwitchExpr(o, ov)
		}
		return nil, fmt.Errorf("switch val type must be a string or object,but %T", val)
	}

	vale, err := parseValueV(str)
	if err != nil {
		return nil, fmt.Errorf("parse switch value error:%w %v", err, str)
	}

	cases, ok := o["case"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("switch case type must be a map,but %v", o)
	}

	sw := &switchExpr{
		val:   vale,
		cases: map[any]exp{},
	}
	for val, exprs := range cases {
		vv, err := parseValueV(val)
		if err != nil {
			return nil, fmt.Errorf("parse switch case value error:%w %v", err, val)
		}
		vvcst, ok := vv.(*constraint)
		if !ok {
			return nil, fmt.Errorf("parse switch cases value is not constraint:%v", val)
		}
		vexpr, err := ParseFromJSONObj(exprs)
		if err != nil {
			return nil, fmt.Errorf("parse switch expression error:%w %v: %v", err, val, exprs)
		}
		sw.cases[vvcst.value] = vexpr

	}

	def := o["default"]
	if def != nil {
		defExpr, err := ParseFromJSONObj(def)
		if err != nil {
			return nil, fmt.Errorf("parse switch default error:%w %v", err, def)
		}
		sw.def = defExpr
	}
	return sw, nil
}

func ParseFromJSONStr(str string) (Expr, error) {
	var o any
	if err := json.Unmarshal([]byte(str), &o); err != nil {
		return nil, err
	}
	return ParseFromJSONObj(o)
}

func ParseFromJSONObj(o any) (Expr, error) {
	switch o := o.(type) {
	case map[string]interface{}:
		for key, val := range o {
			fac := expParserFactory[key]
			if fac != nil {
				e, err := fac(o, val)
				if err != nil {
					return nil, err
				}
				return e, nil
			}
		}
		return nil, fmt.Errorf("not found exp field: %v", o)
	case []any:
		es := &exps{}
		for _, a := range o {
			e, err := ParseFromJSONObj(a)
			if err != nil {
				return nil, err
			}
			es.exps = append(es.exps, e)
		}
		return es, nil
	case string:
		e, err := parseExpr(strings.TrimSpace(o))
		if err != nil {
			return nil, err
		}
		return e, nil
	default:
		return nil, fmt.Errorf("invalid type %T", o)
	}
}

var (
	ParseExpr  = parseExpr
	ParseValue = parseValueV
)

type noneExpr struct {
}

func (n *noneExpr) Exec(c *Context) error {
	return nil
}

func parseExpr(e string) (exp, error) {
	if strings.HasPrefix(e, "#") {
		return &noneExpr{}, nil
	}
	switch {
	case isFuncCall(e):
		v, err := parseValueV(e)
		if err != nil {
			return nil, err
		}
		f, ok := v.(*funcVariable)
		if !ok {
			return nil, errors.New("invalid expression not function:" + e)
		}
		return &callCond{f}, nil
	case isSetCond(e):
		kvs := strings.SplitN(e, "=", 2)
		v, err := parseValueV(kvs[1])
		if err != nil {
			return nil, fmt.Errorf("parse set cond value error:%v", err)
		}
		jp, err := jsonpath.Compile(strings.TrimSpace(kvs[0]))
		if err != nil {
			return nil, fmt.Errorf("parse set cond varname as jsonpath error :%v %w", kvs[0], err)
		}
		return &setCond{
			varName:   strings.TrimSpace(kvs[0]),
			val:       v,
			nameJPath: jp,
		}, nil
	default:
		switch e {
		case "break":
			return &errExpr{err: errBreak}, nil
		case "return":
			return &errExpr{err: errReturn}, nil

		}
		return nil, fmt.Errorf("invalid exp:%s", e)
	}
}

//func parseFuncVal(e string) (val, error) {
//
//}

func parseValueV(e string) (Val, error) {
	tks, err := parseTokenizer(e)
	if err != nil {
		return nil, err
	}
	return parseTokenAsVal(tks)
}

type valueParser struct {
}
type stack[T any] struct {
	data []T
}

func (s *stack[T]) push(data T) {
	s.data = append(s.data, data)
}
func (s *stack[T]) top() T {
	return s.data[len(s.data)-1]
}
func (s *stack[T]) pop() T {
	if len(s.data) == 0 {
		panic("stack is empty")
	}
	data := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return data
}
func (s *stack[T]) empty() bool {
	return len(s.data) == 0
}

type exprStack struct {
	ss  *stack[any]
	tkn []any
}

func newExprStack() *exprStack {
	return &exprStack{
		ss:  new(stack[any]),
		tkn: []any{},
	}
}

var (
	priority = map[byte]int{
		'*': 3,
		'/': 3,
		'+': 1,
		'-': 1,
	}
)

func parseTokenAsVal(tkns []tokenV) (Val, error) {

	vs := &stack[any]{}
	ps := vs
	//temps := stack[any]{}
	for _, tkn := range tkns {
		switch tkn.kind {
		case ')':
			args := stack[Val]{}
		end:
			for !vs.empty() {
				v := vs.pop()
				switch v := v.(type) {
				case Val:
					args.push(v)

				}
				if v == '(' {
					break end
				}
			}
			if vs.empty() {
				return nil, errors.New("invalid func call, no func name")
			}
			v, ok := vs.pop().(*variable)
			if !ok {
				return nil, errors.New("invalid func call, func kind not valid:" + reflect.TypeOf(v).String())
			}
			funName := strings.TrimSpace(v.varName)
			fun := funtables[funName]
			if fun == nil {
				return nil, errors.New("invalid func call, func name not found:" + funName)
			}
			fv := &funcVariable{
				fun: fun,
			}
			for !args.empty() {
				v := args.pop()
				fv.args = append(fv.args, v)
			}
			vs.push(fv)
		case ',':
			if !ps.empty() {
				ts, ok := ps.top().(*stack[any])
				if ok {
					ps.pop()
					ps.push(ts.pop())
					vs = ps
				}
			}
		case 0:
			varName := strings.TrimSpace(tkn.tkn)
			jp, err := jsonpath.Compile(varName)

			if err != nil {
				return nil, fmt.Errorf("invalid var cannot parse as jsonpath:" + varName)
			}

			if v, ok := isVariableConstraint(varName); ok {
				vs.push(&constraint{
					value: v,
				})
			} else {
				vs.push(&variable{
					varName: varName,
					varPath: jp,
				})
			}

		case -1:
			vs.push(&constraint{
				value: tkn.tkn,
			})
		case '(':

			vs.push('(')
		case '+', '-', '*', '/':
			if ps.empty() {
				return nil, fmt.Errorf("invalid '%v' ", tkn.tkn)
			}
			ps.top()
			ts := newExprStack()
			vs.push(ts)

		default:
			panic("invalid token kind")
		}

	}
	//if !temps.empty() {
	//	vs.push(temps.pop())
	//}
	if len(vs.data) != 1 {
		return nil, errors.New("invalid expr ,not completed expr")
	}
	a, ok := vs.pop().(Val)
	if !ok {
		return nil, errors.New("invalid expr")
	}
	return a, nil
}

type tokenV struct {
	tkn  string
	kind int
}

type tokenizer struct {
	next   func(c rune) error
	tokens []tokenV
	tkn    []rune
}

func isVariableConstraint(s string) (any, bool) {
	switch s {
	case "true":
		return true, true
	case "false":
		return false, true
	case "nil":
		return nil, true
	}
	f, err := strconv.ParseFloat(s, 64)
	return f, err == nil

}

func parseTokenizer(exp string) ([]tokenV, error) {
	t := tokenizer{
		tokens: []tokenV{},
	}
	t.next = t.statStart
	for _, r := range exp {
		err := t.next(r)
		if err != nil {
			return nil, err
		}
	}
	if len(t.tkn) > 0 {
		t.tokens = append(t.tokens, tokenV{
			tkn: string(t.tkn),
		})
	}
	return t.tokens, nil

}

func (t *tokenizer) appendToken(kind int) {
	if len(t.tkn) > 0 {
		t.tokens = append(t.tokens, tokenV{
			tkn: string(t.tkn),
		})
	}

	t.tokens = append(t.tokens, tokenV{
		tkn:  string(byte(kind)),
		kind: kind,
	})
	t.tkn = t.tkn[:0]

}

func (t *tokenizer) statStart(r rune) error {
	switch r {
	case '(', ')':
		t.appendToken(int(r))
	case '\'':
		t.next = t.statStringStart
	case ',':
		t.appendToken(',')
	default:
		t.tkn = append(t.tkn, r)
	}
	return nil
}

func (t *tokenizer) statStringStart(r rune) error {
	switch r {
	case '\'':
		t.tokens = append(t.tokens, tokenV{
			tkn:  string(t.tkn),
			kind: -1,
		})
		t.tkn = t.tkn[:0]
		t.next = t.statStart
	case '\\':
		t.next = t.escapeNext(t.statStringStart)
	default:
		t.tkn = append(t.tkn, r)
	}
	return nil
}

func (t *tokenizer) escapeNext(statFunc func(c rune) error) func(c rune) error {
	return func(c rune) error {
		switch c {
		case 'n':
			t.tkn = append(t.tkn, '\n')
		default:
			t.tkn = append(t.tkn, c)
		}
		t.next = statFunc
		return nil
	}
}

type iterator interface {
	getNext() (k any, val any, ok bool)
}

type dataIterator struct {
	start  float64
	end    float64
	offset float64
}

func (d *dataIterator) getNext() (k any, val any, ok bool) {
	hasnext := d.offset < d.end
	k = d.offset
	val = d.offset + d.start
	d.offset++
	return k, val, hasnext
}
