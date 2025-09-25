// sym.go
package ast

import "fmt"

// Node AST 接口
type Node interface {
	Eval(env Env) (float64, error)
	String() string
}

// AST 节点实现

type Number struct{ Val float64 }

func (n *Number) Eval(env Env) (float64, error) { return n.Val, nil }
func (n *Number) String() string                { return fmt.Sprintf("%v", n.Val) }

type Variable struct{ Name string }

func (v *Variable) Eval(env Env) (float64, error) {
	val, ok := env.Vars[v.Name]
	if !ok {
		return 0, fmt.Errorf("undefined variable: %s", v.Name)
	}
	return val, nil
}
func (v *Variable) String() string { return v.Name }

type Unary struct {
	Op string
	X  Node
}

func (u *Unary) Eval(env Env) (float64, error) {
	x, err := u.X.Eval(env)
	if err != nil {
		return 0, err
	}
	switch u.Op {
	case "-":
		return -x, nil
	default:
		return 0, fmt.Errorf("unknown unary op %s", u.Op)
	}
}
func (u *Unary) String() string { return "(" + u.Op + u.X.String() + ")" }

type Binary struct {
	Op   string
	L, R Node
}

func (b *Binary) Eval(env Env) (float64, error) {
	l, err := b.L.Eval(env)
	if err != nil {
		return 0, err
	}
	r, err := b.R.Eval(env)
	if err != nil {
		return 0, err
	}
	switch b.Op {
	case "+":
		return l + r, nil
	case "-":
		return l - r, nil
	case "*":
		return l * r, nil
	case "/":
		return l / r, nil
	case "^":
		// use math.Pow
		return pow(l, r), nil
	default:
		return 0, fmt.Errorf("unknown binary op %s", b.Op)
	}
}
func (b *Binary) String() string { return "(" + b.L.String() + b.Op + b.R.String() + ")" }

type Call struct {
	Name string
	Args []Node
}

func (c *Call) Eval(env Env) (float64, error) {
	fn, ok := env.Funcs[c.Name]
	if !ok {
		return 0, fmt.Errorf("undefined function: %s", c.Name)
	}
	// eval args
	args := make([]float64, 0, len(c.Args))
	for _, a := range c.Args {
		v, err := a.Eval(env)
		if err != nil {
			return 0, err
		}
		args = append(args, v)
	}
	return fn(args)
}
func (c *Call) String() string {
	s := c.Name + "("
	for i, a := range c.Args {
		if i > 0 {
			s += ", "
		}
		s += a.String()
	}
	s += ")"
	return s
}

// helper pow (so we don't need to import math in many files)
func pow(a, b float64) float64 {
	// use math.Pow
	return float64FromMathPow(a, b)
}

// Environment for evaluation
type Env struct {
	Vars  map[string]float64
	Funcs map[string]func([]float64) (float64, error)
}

// yySymType required by goyacc - fields used in grammar actions
type yySymType struct {
	yys     int
	node    Node
	str     string
	strs    []string
	num     float64
	nodes   []Node
	boolean bool
	kv      KV
	kvs     []KV
}

func (s *yySymType) SetStr(v string)  { s.str = v }
func (s *yySymType) SetNum(v float64) { s.num = v }
func (s *yySymType) SetBool(v bool)   { s.boolean = v }

type YySymType = yySymType

func YYParse(lex yyLexer) {
	yyNewParser().Parse(lex)
}

type String struct {
	Val string
}

func (s *String) Eval(env Env) (float64, error) {
	return 0, nil
}

func (s *String) String() string {
	return s.Val
}

type Nil struct {
}

func (n *Nil) Eval(env Env) (float64, error) {
	//TODO implement me
	return 0, nil
}

func (n *Nil) String() string {
	//TODO implement me
	return "nil"
}

type Bool struct {
	Val bool
}

func (b *Bool) Eval(env Env) (float64, error) {
	//TODO implement me
	return 0, nil
}

func (b *Bool) String() string {
	return fmt.Sprintf("%v", b.Val)
}

type Setter interface {
	SetRoot(node Node)
}

type Ternary struct {
}

type Set struct {
	Const bool
	L     Node
	R     Node
}

func (s *Set) Eval(env Env) (float64, error) {
	//TODO implement me
	return 0, nil
}

func (s *Set) String() string {
	//TODO implement me
	return "set"
}

type Access struct {
	L Node
	R Node
}

func (s *Access) Eval(env Env) (float64, error) {
	//TODO implement me
	return 0, nil
}

func (s *Access) String() string {
	//TODO implement me
	return "access"
}

type KV struct {
	K Node
	V Node
}

type MapSet struct {
	Kvs []KV
}

func (m *MapSet) Eval(env Env) (float64, error) {
	//TODO implement me
	return 0, nil
}

func (m *MapSet) String() string {
	//TODO implement me
	return fmt.Sprintf("%v", m.Kvs)
}

type ArrDef struct {
	V []Node
}

func (a *ArrDef) Eval(env Env) (float64, error) {
	//TODO implement me
	return 0, nil
}

func (a *ArrDef) String() string {
	//TODO implement me
	return "arr"
}

type ArrAccess struct {
	L, R Node
}

func (a *ArrAccess) Eval(env Env) (float64, error) {
	//TODO implement me
	return 0, nil
}

func (a *ArrAccess) String() string {
	//TODO implement me
	return "arrAccess"
}

type SliceCut struct {
	V  Node
	St Node
	Ed Node
}

func (s *SliceCut) Eval(env Env) (float64, error) {
	//TODO implement me
	return 0, nil
}

func (s *SliceCut) String() string {
	//TODO implement me
	return "sliceCut"
}

type Lambda struct {
	L []string
	R Node
}

func (l *Lambda) Eval(env Env) (float64, error) {
	//TODO implement me
	return 0, nil
}

func (l *Lambda) String() string {
	//TODO implement me
	return "lambda"
}
