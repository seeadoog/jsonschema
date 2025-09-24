package expr

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/seeadoog/jsonschema/v2/expr/ast"
	"github.com/seeadoog/jsonschema/v2/jsonpath"
	"regexp"
	"strconv"
	"strings"
)

/*
"if":"eq(name,5)"

*/

type Context struct {
	table                   map[string]any
	funcs                   map[string]ScriptFunc
	returnVal               []any
	IgnoreFuncNotFoundError bool
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

func (c *Context) GetByJp(key string) any {
	jp, err := jsonpath.Compile(key)
	if err != nil {
		return nil
	}
	return c.GetJP(jp)
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
		if !strings.HasPrefix(key, "$") {
			panic(fmt.Sprintf("func '%s' not registerd by RegisterDynamicFunc", key))
		}
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
		re, ok := err.(*Return)
		if ok {
			c.returnVal = re.Var
			return nil
		}
	}
	return err
}

func (c *Context) GetReturn() []any {
	return c.returnVal
}

func (c *Context) GetTable() map[string]any {
	return c.table
}

type setValue struct {
	key Val
	jp  *jsonpath.Complied
	val Val
}

// a.b = 5   accs
func (s *setValue) Val(c *Context) any {
	//v := s.val.Val(c)
	//if s.jp == nil {
	//	c.Set(s.key, v)
	//} else {
	//	c.SetJP(s.jp, v)
	//}
	//return v
	return s.Set(c, s.val.Val(c))
}

func setFor(c *Context, left Val, v any) {
	switch vs := (left).(type) {
	case *accessVal:
		parent, ok := vs.left.Val(c).(map[string]interface{})
		if !ok {
			set, ok := vs.left.(setter)
			if ok {
				parent = map[string]any{}
				set.Set(c, parent)
			} else {
				return
			}
			//return v
		}
		varn, ok := vs.right.(*variable)
		if !ok {
			return
		}
		parent[varn.varName] = v
	case *variable:
		vs.Set(c, v)
		return
	case *arrAccessVal:
		rv := vs.right.Val(c)
		switch rvv := rv.(type) {
		case float64:
			parent, ok := vs.left.Val(c).([]any)

			idx := int(rvv)
			if !ok {
				if parent != nil {
					return
				}
				parent = make([]any, idx+1)
				set, ok := vs.left.(setter)
				if ok {
					set.Set(c, parent)
				}
			} else {
				if len(parent) <= idx {
					old := parent
					parent = make([]any, idx+1)
					copy(parent, old)
					set, ok := vs.left.(setter)
					if ok {
						set.Set(c, parent)
					}
				}
			}
			parent[idx] = v

		case string:
			parent, ok := vs.left.Val(c).(map[string]interface{})
			if !ok {
				set, ok := vs.left.(setter)
				if ok {
					parent = map[string]any{}
					set.Set(c, parent)
				} else {
					return
				}
				//return v
			}
			parent[rvv] = v
		}

	}
}

func (s *setValue) Set(c *Context, val any) any {
	setFor(c, s.key, val)
	return val
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
	if v.varPath == nil {
		return c.Get(v.varName)
	}
	return c.GetJP(v.varPath)
}

func (v *variable) Set(c *Context, val any) any {
	if v.varPath == nil {
		c.Set(v.varName, val)
		return val
	}
	c.SetJP(v.varPath, val)
	return val
}

type constraint struct {
	value any
}

func (c *constraint) Val(ctx *Context) any {
	return c.value
}

type ScriptFunc func(ctx *Context, args ...Val) any
type funcVariable struct {
	funcName string
	fun      func(ctx *Context, args ...Val) any
	args     []Val
}

func (c *funcVariable) Val(ctx *Context) any {
	if c.fun == nil {
		if ctx.funcs != nil {
			f := ctx.funcs[c.funcName]
			if f != nil {
				return f(ctx, c.args...)
			} else {
				return newErrorf("function '%s' not found in table", c.funcName)
			}
		}
	}
	return c.fun(ctx, c.args...)
}

type ifCond struct {
	cond     Val
	thenCond exp
	elseCond exp
}

func (i *ifCond) Exec(c *Context) error {
	if BoolCond(i.cond.Val(c)) {
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

type valCond struct {
	val Val
}

func convertToError(o any) error {
	switch o := o.(type) {
	case *Return:
		return o
	case *Error:
		return o
	}
	return nil
}

func (v *valCond) Exec(c *Context) error {
	o := v.val.Val(c)
	return convertToError(o)
}

func (s *setCond) Exec(c *Context) error {
	if s.nameJPath == nil {
		c.Set(s.varName, s.val.Val(c))
		return nil
	}
	return c.SetJP(s.nameJPath, s.val.Val(c))
}

type callCond struct {
	fun *funcVariable
}

type Error struct {
	Err string
}

func (e *Error) Error() string {
	return e.Err
}
func newErrorf(format string, args ...interface{}) *Error {
	return &Error{Err: fmt.Sprintf(format, args...)}
}

func (c *callCond) Exec(ctx *Context) error {
	o := c.fun.Val(ctx)
	return convertToError(o)
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
	setCondReg = regexp.MustCompile(`([$_\-0-9a-zA-Z.\[\]]+\s*)=([^=])\s*(.*)`)

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

type switchCasesExpr struct {
	cases       []casesExprs
	defaultExpr Expr
}

func (s *switchCasesExpr) Exec(c *Context) error {
	for _, exprs := range s.cases {
		if BoolCond(exprs.val.Val(c)) {
			return exprs.expr.Exec(c)
		}
	}
	if s.defaultExpr != nil {
		return s.defaultExpr.Exec(c)
	}
	return nil
}

func parseSwitchExpr(o map[string]any, val map[string]any) (exp, error) {
	switchCases := &switchCasesExpr{}
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
		return nil, fmt.Errorf("parse from json str error:%w", err)
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
			_, ok := e.(*noneExpr)
			if ok {
				continue
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

func isJsonPath(s string) bool {
	return strings.Contains(s, ".") || strings.Contains(s, "[")
}
func parseExpr(e string) (exp, error) {
	if strings.HasPrefix(e, "#") {
		return &noneExpr{}, nil
	}
	switch {
	default:
		switch e {
		case "break":
			return &errExpr{err: errBreak}, nil
		case "return":
			return &errExpr{err: errReturn}, nil
		}
		v, err := parseValueV(e)
		if err != nil {
			return nil, fmt.Errorf("parse stmt as value error:%w :: %s", err, e)
		}
		return &valCond{
			val: v,
		}, nil
		//return nil, fmt.Errorf("invalid exp:%s", e)
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
	lex := &lexer{
		tokens: tks,
	}
	ast.YYParse(lex)
	if lex.err != nil {
		return nil, fmt.Errorf("parse value error:%v ,%v", lex.err, e)
	}
	v, err := ParseValueFromNode(lex.root, false)
	if err != nil {
		return nil, fmt.Errorf("parse value error:%w ,%v", err, e)
	}
	return v, nil
	//return parseTokenAsVal(tks)
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

type tokenV struct {
	tkn  string
	kind int
}

type tokenizer struct {
	next   func(c rune) error
	tokens []tokenV
	tkn    []rune
	exp    []rune
	pos    int
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
		exp:    []rune(exp),
	}
	t.next = t.statStart
	r := []rune(exp)
	for t.pos = 0; t.pos < len(r); t.pos++ {
		err := t.next(r[t.pos])
		if err != nil {
			return nil, fmt.Errorf("parse exp error as token error:%w '%v'", err, exp)
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
			tkn:  string(t.tkn),
			kind: variables,
		})
	}

	t.tokens = append(t.tokens, tokenV{
		tkn:  string(byte(kind)),
		kind: kind,
	})
	t.tkn = t.tkn[:0]

}

func (t *tokenizer) getNext() (rune, bool) {
	t.pos++
	if t.pos >= len(t.exp) {
		return 0, false
	}
	return t.exp[t.pos], true
}

func (t *tokenizer) pre() (rune, bool) {
	if t.pos-1 < 0 {
		return 0, false
	}
	return t.exp[t.pos-1], true
}
func (t *tokenizer) appendId() {
	if len(t.tkn) > 0 {
		seg := string(t.tkn)
		kind := variables
		switch seg {
		case "or":
			kind = ast.ORR
		case "const":
			kind = ast.CONST
		}
		t.tokens = append(t.tokens, tokenV{
			tkn:  seg,
			kind: kind,
		})
		t.tkn = t.tkn[:0]
	}
}

func (t *tokenizer) statStart(r rune) error {
	switch r {
	case '(', ')', '?', ';', '{', '}', '[', ']':
		t.appendToken(int(r))
	case '#':
		t.next = func(c rune) error {
			return nil
		}
	case ':':
		c, ok := t.getNext()
		if !ok {
			return fmt.Errorf("unexpected  eof after ':'")
		}
		if c == ':' {
			t.appendToken(ast.ACC)
			return nil
		}
		t.pos--
		t.appendToken(int(r))
	case '\'':
		t.next = t.statStringStart
	case '`':
		t.next = t.statStringStartWith('`')
	case '"':
		t.next = t.statStringStartWith('"')
	case ',':
		t.appendToken(',')
	case ' ', '\t', '\n':
		t.appendId()

	case '+', '*', '/', '^':
		t.appendToken(int(r))

	case '-':
		c, ok := t.getNext()
		if !ok {
			return fmt.Errorf("unexpected  eof after '-'")
		}
		if c == '>' {
			t.appendToken(ast.ACC)
			return nil
		}
		t.pos--
		t.appendToken(int(r))
	case '!':
		c, ok := t.getNext()
		if !ok {
			return fmt.Errorf("unexpected  eof after '!'")
		}
		if c == '=' {
			t.appendToken(ast.NOTEQ)
			return nil
		}
		t.pos--
		t.appendToken(int(r))
	case '=':
		t.next = t.statParseEq
	case '|':
		t.next = t.statParseOr
	case '&':
		t.next = t.statParseAND
	case '>':
		c, ok := t.getNext()
		if !ok {
			return fmt.Errorf("unexpected  eof after '>'")
		}
		if c == '=' {
			t.appendToken(ast.GTE)
			return nil
		}
		t.pos--
		t.appendToken(ast.GT)
	case '<':
		c, ok := t.getNext()
		if !ok {
			return fmt.Errorf("unexpected  eof after '>'")
		}
		if c == '=' {
			t.appendToken(ast.LTE)
			return nil
		}
		t.pos--
		t.appendToken(ast.LT)
	default:
		t.tkn = append(t.tkn, r)
	}
	return nil
}

func (t *tokenizer) statParseEq(r rune) error {
	if r != '=' {
		t.appendToken('=')
		t.pos--
		t.next = t.statStart
		return nil
	}
	t.appendToken(ast.EQ)
	t.next = t.statStart
	return nil
}
func (t *tokenizer) statParseAND(r rune) error {
	if r != '&' {
		return errors.New("invalid token after & ")
	}
	t.appendToken(ast.AND)
	t.next = t.statStart
	return nil
}
func (t *tokenizer) statParseOr(r rune) error {
	if r != '|' {
		return errors.New("invalid token after | ")
	}
	t.appendToken(ast.OR)
	t.next = t.statStart
	return nil
}

func (t *tokenizer) statStringStart(r rune) error {
	switch r {
	case '\'':
		t.tokens = append(t.tokens, tokenV{
			tkn:  string(t.tkn),
			kind: constant,
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
func (t *tokenizer) statStringStartWith(c rune) func(c rune) error {
	var fff func(rune) error
	fff = func(r rune) error {
		switch r {
		case c:
			t.tokens = append(t.tokens, tokenV{
				tkn:  string(t.tkn),
				kind: constant,
			})
			t.tkn = t.tkn[:0]
			t.next = t.statStart
		case '\\':
			t.next = t.escapeNext(fff)
		default:
			t.tkn = append(t.tkn, r)
		}
		return nil
	}
	return fff
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

type Return struct {
	Var []any
}

func (r *Return) Error() string {
	return fmt.Sprintf("return: %v", r.Var)
}

func ValueOfReturn(e error) []any {
	r, ok := e.(*Return)
	if ok {
		return r.Var
	}
	return nil
}
