package main

import (
	"fmt"
	"github.com/expr-lang/expr"
)

func main() {
	env := map[string]interface{}{
		"greet":   "Hello, %v!",
		"age":     "xx",
		"names":   []string{"world", "you"},
		"sprintf": fmt.Sprintf,
	}
	// ass::filter(e => e.name > 5)
	code := `greet + age`

	program, err := expr.Compile(code, expr.Env(env))
	if err != nil {
		panic(err)
	}

	output, err := expr.Run(program, env)
	if err != nil {
		panic(err)
	}

	fmt.Println(output)
}
