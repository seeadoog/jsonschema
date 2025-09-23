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
	}
	env2["set"] = func(k string, v any) any {
		env2[k] = v
		return k
	}
	// ass::filter(e => e.name > 5)
	code := `age = 5`
	b.ReportAllocs()
	program, err := expr.Compile(code)
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
	fmt.Println("start") // define('map_to_str',for($1))
	e, err := expr2.ParseValue("arr[2:]")
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()
	tb := map[string]interface{}{
		"status": float64(2000000000),
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
	fmt.Println(reflect.TypeOf(e.Val(vm)))

	fmt.Println("result:", e.Val(vm))
	fmt.Println(tb)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Val(vm)
	}
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
	//a := map[string]interface{}{}
	//b.ReportAllocs()
	//for i := 0; i < b.N; i++ {
	//
	//}
}
