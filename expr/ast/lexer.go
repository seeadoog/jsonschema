// lexer.go
package ast

import (
	"fmt"
	"strconv"
	"unicode"
)

type lexer struct {
	input    string
	pos      int
	rootNode Node
	errors   []string
}

func (l *lexer) Lex(lval *yySymType) int {
	// skip spaces
	for l.pos < len(l.input) && unicode.IsSpace(rune(l.input[l.pos])) {
		l.pos++
	}
	if l.pos >= len(l.input) {
		return 0 // EOF
	}
	ch := l.input[l.pos]

	// single-char tokens
	switch ch {
	case '+', '-', '*', '/', '(', ')', ',', '^':
		l.pos++
		switch ch {
		case '+':
			return int('+')
		case '-':
			return int('-')
		case '*':
			return int('*')
		case '/':
			return int('/')
		case '(':
			return int('(')
		case ')':
			return int(')')
		case ',':
			return int(',')
		case '^':
			return int('^')
		case '&':
			return int('&')
		case '!':
			return int('!')
		case '|':
			return int('|')
		}
	}
	if ch == '\'' {
		l.pos++
		start := l.pos
		for l.pos < len(l.input) {
			cc := l.input[l.pos]
			if cc == '\'' {
				break
			}
			if cc == '\\' {
				if l.pos+1 >= len(l.input) {
					l.Error("invalid escape sequence")
					return 0
				}
			}
			l.pos++
		}
		lval.str = l.input[start:l.pos]
		l.pos++
		return STRING
	}
	// number (integer or float)
	if (ch >= '0' && ch <= '9') || ch == '.' {
		start := l.pos
		dotSeen := false
		for l.pos < len(l.input) {
			c := l.input[l.pos]
			if c == '.' {
				if dotSeen {
					break
				}
				dotSeen = true
			} else if c < '0' || c > '9' {
				break
			}
			l.pos++
		}
		s := l.input[start:l.pos]
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			// fallback
			v = 0
		}
		lval.num = v
		return NUMBER
	}

	// identifier: letter or underscore, then letters/digits/underscore
	if unicode.IsLetter(rune(ch)) || ch == '_' {
		start := l.pos
		l.pos++
		for l.pos < len(l.input) {
			c := l.input[l.pos]
			if !(unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) || c == '_') {
				break
			}
			l.pos++
		}
		s := l.input[start:l.pos]
		lval.str = s
		return IDENT
	}

	// unknown
	l.pos++
	return int(ch)
}

func (l *lexer) Error(e string) {
	fmt.Println("lexer error:", e)
	l.errors = append(l.errors, e)
}
