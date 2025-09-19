package expr

import (
	"fmt"
	"github.com/seeadoog/jsonschema/v2/expr/ast"
	"github.com/seeadoog/jsonschema/v2/jsonpath"
	"strconv"
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
	l.err = append(l.err, s)
}

func ParseValueFromNode(node ast.Node) (Val, error) {
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
		var jp *jsonpath.Complied
		if isJsonPath(n.Name) {
			var err error
			jp, err = jsonpath.Compile(n.Name)
			if err != nil {
				return nil, fmt.Errorf("parse vname as jsonpath error:%w", err)
			}
		}

		return &variable{
			varName: n.Name,
			varPath: jp,
		}, nil
	case *ast.Call:
		fun := funtables[n.Name]
		if fun == nil {
			return nil, fmt.Errorf("func '%s' is not defined", n.Name)
		}
		args := make([]Val, 0, len(n.Args))
		for _, arg := range n.Args {
			argv, err := ParseValueFromNode(arg)
			if err != nil {
				return nil, err
			}
			args = append(args, argv)
		}
		return &funcVariable{
			fun:  fun,
			args: args,
		}, nil
	case *ast.Unary:
		val, err := ParseValueFromNode(n.X)
		if err != nil {
			return nil, fmt.Errorf("unary parse val error:%w", err)
		}
		switch n.Op {
		case "!":
			return &funcVariable{
				fun:  notFunc,
				args: []Val{val},
			}, nil
		case "-":
			return &funcVariable{
				fun:  negativeFunc,
				args: []Val{val},
			}, nil
		}
		return nil, fmt.Errorf("unknown unary operator:%s", n.Op)
	case *ast.Binary:
		lv, err := ParseValueFromNode(n.L)
		if err != nil {
			return nil, fmt.Errorf("binary parse val L error:%w %s", err, lv)
		}
		rv, err := ParseValueFromNode(n.R)
		if err != nil {
			return nil, fmt.Errorf("binary parse val R error:%w %s", err, lv)
		}
		var fun ScriptFunc
		switch n.Op {
		case "+":
			fun = addFunc
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
		case "||":
			fun = orFunc
		case "==":
			fun = eqFunc
		case "<":
			fun = lessFunc
		case "<=":
			fun = lessOrEqual
		case ">":
			fun = largeFunc
		case ">=":
			fun = largeOrEqual
		case "!=":
			fun = notEqFunc
		default:
			return nil, fmt.Errorf("unknown operator of binary :%s %s", n.Op, n)
		}
		return &funcVariable{
			fun:  fun,
			args: []Val{lv, rv},
		}, nil
	case *ast.Set:
		var jp *jsonpath.Complied
		var err error
		if isJsonPath(n.L) {
			jp, err = jsonpath.Compile(n.L)
			if err != nil {
				return nil, fmt.Errorf("parse set field error:%w", err)
			}
		}

		val, err := ParseValueFromNode(n.R)
		if err != nil {
			return nil, fmt.Errorf("set parse val error:%w", err)
		}
		return &setValue{
			key: n.L,
			jp:  jp,
			val: val,
		}, nil
	default:
		return nil, fmt.Errorf("invalid ast.Node type :%T", node)
	}

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
	//arr := arrPool.Get().([]string)
	arr := make([]string, 0, len(s.vals))
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
	//arrPool.Put(arr[:0])
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
