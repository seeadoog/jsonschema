package main

import (
	"encoding/json"
	"fmt"
	"github.com/expr-lang/expr"
	expr2 "github.com/seeadoog/jsonschema/v2/expr"
	"strings"
	"testing"
)

func BenchmarkExpr(b *testing.B) {

	var i int
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
		"hls": func(c int) int {
			i++
			return i
		},
		"usr": &User{
			Name: "abc",
			Age:  0,
			Chd:  nil,
			Arr:  nil,
		},
	}
	env2["set"] = func(k string, v any) any {
		env2[k] = v
		return k
	}
	// ass::filter(e => e.name > 5)
	code := `hls(1)`
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

		return nil
	}, 0)
	e, err := expr2.ParseValue(`
status == '3'
`)
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()
	tb := map[string]interface{}{
		"status": "3",
		"doc":    map[string]any{},
		"json": map[string]any{
			"data":  "hello",
			"text":  "js is ok",
			"text2": "js is ok",
			"text3": "js is ok",
			"arr":   []any{1.0, 2.2, 3.3},
			"json":  map[string]any{},
		},
		"usr": &User{
			Name: "55",
			Age:  18,
			Chd: &User{
				Name: "chd",
				Age:  3,
				Chd:  nil,
			},
		},
	}
	vm := expr2.NewContext(tb)
	vm.ForceType = false

	vm.SetFunc("set_self", expr2.FuncDefine(func() any {
		//tb[a] = b
		return nil
	}))

	fmt.Println("result:", e.Val(vm))
	printJson(tb)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Val(vm)
	}

	fmt.Println("call_num:", i, e.Val(vm))
}

func printJson(v any) {
	bs, _ := json.MarshalIndent(v, "", " ")
	fmt.Println(string(bs))
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

type User struct {
	Name string
	Age  int
	Chd  *User
	Arr  []int
}

func TestExpr(t *testing.T) {
	e, err := expr2.ParseValue(`
usr->Name
`)
	if err != nil {
		panic(err)
	}
	c := expr2.NewContext(map[string]interface{}{
		"usr": &User{
			Name: "55",
			Age:  18,
			Chd: &User{
				Name: "chd",
				Arr:  []int{1, 2, 3},
			},
			Arr: []int{1, 2, 3},
		},
		"arr": []int{1, 3, 4},
		"json": map[string]interface{}{
			"data": "hello world",
			"text": "js is ok",
		},
		"sub": "ist",
		"cha": "2",
	})
	c.NewCallEnv = true
	c.ForceType = true
	fmt.Println("result:", e.Val(c))

	bs, _ := json.MarshalIndent(c.GetTable(), "", "  ")
	fmt.Println(string(bs))

}
