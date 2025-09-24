package main

import (
	"fmt"
	"github.com/expr-lang/expr"
	expr2 "github.com/seeadoog/jsonschema/v2/expr"
	"reflect"
	"strings"
	"testing"
)

func BenchmarkExpr(b *testing.B) {

	env2 := map[string]interface{}{
		"greet":   "Hello, %v!",
		"age":     "xx",
		"d":       "xx",
		"names":   []string{"world", "you"},
		"sprintf": fmt.Sprintf,
		"a":       1,
		"b":       2,
		"status":  1,
		"obj": map[string]any{
			"hello": "world",
		},
	}
	env2["set"] = func(k string, v any) any {
		env2[k] = v
		return k
	}
	// ass::filter(e => e.name > 5)
	code := `[1,2,3,4,a]`
	b.ReportAllocs()
	program, err := expr.Compile(code)
	if err != nil {
		panic(err)
	}
	n, err := expr.Run(program, env2)
	fmt.Println(n)
	if err != nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {

		expr.Run(program, env2)

	}
}

func BenchmarkEpr(b *testing.B) {
	fmt.Println("start") // define('map_to_str',for($1))
	expr2.RegisterDynamicFunc("set_self", 0)

	i := 0

	expr2.RegisterFunc("hls", func(ctx *expr2.Context, args ...expr2.Val) any {
		i++
		return nil
	}, 0)
	e, err := expr2.ParseValue(``)
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()
	tb := map[string]interface{}{
		"status": float64(1),
		"doc":    map[string]any{},
		"json": map[string]any{
			"data":  "hello world",
			"text":  "js is ok",
			"text2": "js is ok",
			"text3": "js is ok",
			"arr":   []any{1.0, 2.2, 3.3},
			"json":  map[string]any{},
		},
		"arr": []any{1.0, 124.0, 125.0, 146.0},
	}
	vm := expr2.NewContext(tb)

	vm.SetFunc("set_self", expr2.FuncDefine(func() any {
		//tb[a] = b
		return nil
	}))
	fmt.Println(reflect.TypeOf(e.Val(vm)))

	fmt.Println("result:", e.Val(vm))
	fmt.Println(tb)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Val(vm)
	}

	fmt.Println("call_num:", i, e.Val(vm))
}

func rawMAP(tb map[string]interface{}) string {

	arr := make([]string, 0)
	for key, val := range tb["json"].(map[string]interface{}) {
		arr = append(arr, fmt.Sprintf("%s=%s", key, val))
	}
	return strings.Join(arr, ";")
}

func BenchmarkRaow(b *testing.B) {
	tb := map[string]interface{}{
		"status": float64(2000000000),
		"json": map[string]any{
			"data":  "hello world",
			"text":  "js is ok",
			"text2": "js is ok",
		},
		"arr": []any{124.0, 125.0, 146.0},
	}
	for i := 0; i < b.N; i++ {
		rawMAP(tb)
	}
}

func BenchmarkIndexer(b *testing.B) {

}
