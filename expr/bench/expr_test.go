package main

import (
	"fmt"
	"github.com/expr-lang/expr"
	expr2 "github.com/seeadoog/jsonschema/v2/expr"
	"testing"
)

func BenchmarkExpr(b *testing.B) {

	env2 := map[string]interface{}{
		"greet":   "Hello, %v!",
		"age":     "xx",
		"d":       "xx",
		"names":   []string{"world", "you"},
		"sprintf": fmt.Sprintf,
	}
	env2["set"] = func(k string, v any) any {
		env2[k] = v
		return k
	}
	// ass::filter(e => e.name > 5)
	code := `age == 'xx' && d == 'xx'`
	b.ReportAllocs()
	program, err := expr.Compile(code, expr.Env(env2))
	if err != nil {
		panic(err)
	}
	n, err := expr.Run(program, env2)
	fmt.Println(n)
	for i := 0; i < b.N; i++ {

		_, err = expr.Run(program, env2)
		if err != nil {
			panic(err)
		}

	}
}
func BenchmarkEpr(b *testing.B) {
	e, err := expr2.ParseValue("age == 'xx' && d == 'xx' ")
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()
	tb := map[string]interface{}{
		"greet": "Hello, %v!",
		"age":   "xx",
		"d":     "xx",
	}
	tb["$$"] = tb
	vm := expr2.NewContext(tb)
	fmt.Println(e.Val(vm))
	for i := 0; i < b.N; i++ {
		e.Val(vm)
	}
}
