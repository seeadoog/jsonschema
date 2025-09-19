// main.go
package ast

import (
	"fmt"
)

func parse(input string) (Node, error) {
	l := &lexer{input: input}
	// call parse (goyacc generates yyParse)
	if yyParse(l) != 0 {
		return nil, fmt.Errorf("parse error")
	}
	// parser sets a global/return value? In our grammar we set yyVAL.node for Input,
	// but to retrieve parse result, we can modify grammar to store result in a package-level var
	// Simpler: after parse, we will expect the lexer or parser to have set a global 'parsed' var.
	// For clarity, let's assume we added a package-level 'lastNode' that grammar writes.
	return lastNode, nil
}

var lastNode Node

// To capture result inside grammar, modify expr.y's top rule action to set lastNode.
// e.g., Input: Expr { lastNode = yyS[yypt-0].node }

func main() {
	// Examples
	tests := []string{
		"sum(1+2*3, sum(1+1),'hello',adf(123,bb,'world'),a&b,!a,a|b(),!a(),a & !b)",
		"sum(1, 2+3, max(4,5),a & b,!a)", // if you add sum into defaultEnv
		"-a * (b + 2)",
		"sqrt(4) + pi",
		"outer(inner(1,2), 3)",
	}

	// prepare env
	env := defaultEnv()
	env.Vars["a"] = 2
	env.Vars["b"] = 5
	// add a "sum" function
	env.Funcs["sum"] = func(args []float64) (float64, error) {
		s := 0.0
		for _, v := range args {
			s += v
		}
		return s, nil
	}
	env.Funcs["outer"] = func(args []float64) (float64, error) {
		if len(args) != 2 {
			return 0, fmt.Errorf("outer expects 2 args")
		}
		return args[0] + args[1], nil
	}
	env.Funcs["inner"] = func(args []float64) (float64, error) {
		if len(args) != 2 {
			return 0, fmt.Errorf("inner expects 2 args")
		}
		return args[0] * args[1], nil
	}

	for _, t := range tests {
		// parse
		l := &lexer{input: t}
		if yyParse(l) != 0 {
			fmt.Println("parse failed:", t)
			continue
		}
		if lastNode == nil {
			fmt.Println("no node produced for:", t)
			continue
		}
		fmt.Println("AST:", lastNode.String())
		v, err := lastNode.Eval(env)
		if err != nil {
			fmt.Println("eval err:", err)
		} else {
			fmt.Println("value:", v)
		}
		fmt.Println("----")
	}
}
