// sym.go
package ast

import "fmt"

func init() {
	yyErrorVerbose = true

	yyDebug = 0
}

// Node AST 接口
type Node interface {
	String() string
}

// AST 节点实现

type Number struct{ Val float64 }

func (n *Number) String() string { return fmt.Sprintf("%v", n.Val) }

type Variable struct{ Name string }

func (v *Variable) String() string { return v.Name }

type Unary struct {
	Op string
	X  Node
}

func (u *Unary) String() string { return "(" + u.Op + u.X.String() + ")" }

type Binary struct {
	Op   string
	L, R Node
}

func (b *Binary) String() string { return "(" + b.L.String() + b.Op + b.R.String() + ")" }

type Call struct {
	Name string
	Args []Node
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
	x, y    int
}

func (s *yySymType) SetStr(v string)  { s.str = v }
func (s *yySymType) SetNum(v float64) { s.num = v }
func (s *yySymType) SetBool(v bool)   { s.boolean = v }
func (s *yySymType) SetPos(x, y int) {
	s.x = x
	s.y = y
}

type YySymType = yySymType

func YYParse(lex yyLexer) {
	yyNewParser().Parse(lex)
}

type String struct {
	Val string
}

func (s *String) String() string {
	return s.Val
}

type Nil struct {
}

func (n *Nil) String() string {
	//TODO implement me
	return "nil"
}

type Bool struct {
	Val bool
}

func (b *Bool) String() string {
	return fmt.Sprintf("%v", b.Val)
}

type Setter interface {
	SetRoot(node Node)
}

type Set struct {
	Const bool
	L     Node
	R     Node
}

func (s *Set) String() string {
	//TODO implement me
	return "set"
}

type Access struct {
	L Node
	R Node
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

func (m *MapSet) String() string {
	//TODO implement me
	return fmt.Sprintf("%v", m.Kvs)
}

type ArrDef struct {
	V []Node
}

func (a *ArrDef) String() string {
	//TODO implement me
	return "arr"
}

type ArrAccess struct {
	L, R Node
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

func (s *SliceCut) String() string {
	//TODO implement me
	return "sliceCut"
}

type Lambda struct {
	L []string
	R Node
}

func (l *Lambda) String() string {
	//TODO implement me
	return "lambda"
}

type Ternary struct {
	C Node
	L Node
	R Node
}

func (t *Ternary) String() string {
	return "ternary"
}

type Const struct {
	L Node
}

func (c *Const) String() string {
	//TODO implement me
	return "const"
}

type NotNil struct {
	N Node
}

func (n *NotNil) String() string {
	return "not nil"
}
