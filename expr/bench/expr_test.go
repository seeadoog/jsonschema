package main

import (
	"encoding/json"
	"fmt"
	"github.com/expr-lang/expr"
	expr2 "github.com/seeadoog/jsonschema/v2/expr"
	"github.com/seeadoog/jsonschema/v2/jsonpath"
	"os"
	"strings"
	"testing"
	"time"
	"unsafe"
)

func BenchmarkExpr(b *testing.B) {

	env2 := map[string]interface{}{
		"status": 3,
		"datad":  "1",

		"doc": map[string]any{},
		"json": map[string]any{
			"data":  "hello",
			"text":  "js is ok",
			"text2": "js is ok",
			"text3": "js is ok",
			"text4": "js is ok",
			"arr":   []any{1.0, 2.2, 3.3},
			"json": map[string]any{
				"xx": 1,
				"x2": map[string]any{
					"a": 1,
				},
			},
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
	env2["set"] = func(k string, v any) any {
		env2[k] = v
		return k
	}
	// ass::filter(e => e.name > 5)
	code := `datad.has_prefix(a)`
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

func eq(m map[string]interface{}, k string, v any) bool {
	return m[k] == v
}

func BenchmarkEpr(b *testing.B) {
	fmt.Println("start") // define('map_to_str',for($1))
	expr2.RegisterDynamicFunc("set_self", 0)

	//f, err := os.Create("bench.pprof")
	//if err != nil {
	//	panic(err)
	//}
	//defer f.Close()
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	i := 0

	expr2.RegisterFunc("hls", func(ctx *expr2.Context, args ...expr2.Val) any {

		return nil
	}, 0)
	e, err := expr2.ParseValue(`
'11'.has_prefix('15')
`)
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()
	tb := map[string]interface{}{
		"status": 3.0,
		"res": &expr2.Result{
			Err:  "err",
			Data: "data",
		},
		"datad": "1",
		"oop": map[string]any{
			"data":  "hello",
			"text":  "js is ok",
			"text2": "js is ok",
			"text3": "js is ok",
			"text4": "js is ok",
		},
		"doc": map[string]any{},
		"o1": map[string]any{
			"data":  "hello",
			"text":  "js is ok",
			"text2": "js is ok",
			"text3": "js is ok",
			"text4": "js is ok",
			"arr":   []any{1.0, 2.2, 3.3},
			"o2": map[string]any{
				"xx": 1,
				"o3": map[string]any{
					"o4": 1,
				},
			},
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
	vm.NewCallEnv = false

	vm.SetFunc("set_self", expr2.FuncDefine(func() any {
		//tb[a] = b
		return nil
	}))

	fmt.Println("result:", e.Val(vm))
	printJson(tb)
	b.ResetTimer()
	var rr bool
	for i := 0; i < b.N; i++ {
		e.Val(vm)
		//mapCP(tb["json"].(map[string]any), tb["json"].(map[string]any))
		//rr = eq(tb, "status", 3)
	}

	fmt.Println("call_num:", i, e.Val(vm), rr)
	fmt.Println(tb)
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

func (u *User) SetField(ctx *expr2.Context, name string, val any) {

	switch name {
	case "name":
		u.Name = val.(string)
	case "age":
		u.Age = int(expr2.NumberOf(val))
	}
}

func (u *User) GetField(c *expr2.Context, key string) any {
	switch key {
	case "name":
		return u.Name
	case "age":
		return u.Age
	case "chd":
		return u.Chd
	case "arr":
		return u.Arr
	}
	return nil
}

func TestExpr(t *testing.T) {
	e, err := expr2.ParseValue(`
nil.has_prefix('15')
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

func mapCP(src, dst map[string]interface{}) {
	for _, i := range src {
		_ = i
	}
}

func BenchmarkJp(b *testing.B) {
	tb := map[string]interface{}{
		"status": 3,
		"datad":  "1",
		"doc":    map[string]any{},
		"json": map[string]any{
			"data":  "hello",
			"text":  "js is ok",
			"text2": "js is ok",
			"text3": "js is ok",
			"text4": "js is ok",
			"arr":   []any{1.0, 2.2, 3.3},
			"json": map[string]any{
				"xx": 1,
				"x2": map[string]any{
					"a": 1,
				},
			},
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
	cp, err := jsonpath.Compile("json.json.x2")
	if err != nil {
		panic(err)
	}
	fmt.Println(cp.Get(tb))
	for i := 0; i < b.N; i++ {
		cp.Get(tb)
	}
}

type V2 interface {
	GetV() any
}

type v2IMp struct {
}

func (v *v2IMp) GetV() any {
	//TODO implement me
	return nil
}

func initV() V2 {
	if os.Getenv("xx") == "3" {
		return nil
	}
	return &v2IMp{}
}

type v4 struct {
}

func (v *v4) Val(c *expr2.Context) any {
	//TODO implement me
	return nil
}

func initV4() expr2.Val {
	if os.Getenv("xx") == "3" {
		return nil
	}
	return &v4{}
}

func BenchmarkInterface(b *testing.B) {
	a := initV()
	for i := 0; i < b.N; i++ {
		a.GetV()
	}
}
func BenchmarkInterface2(b *testing.B) {
	a := initV4()
	for i := 0; i < b.N; i++ {
		a.Val(nil)
	}
}

func printAddr(p *int) {
	fmt.Println(uintptr(unsafe.Pointer(p)))
}

func printAddrOff(p *int, b *int) {
	fmt.Println(uintptr(unsafe.Pointer(p)) - uintptr(unsafe.Pointer(b)))
}

func TestName(t *testing.T) {

	a := 1
	b := 2
	c := add(a, b)

	printAddr(&a)
	printAddr(&b)
	printAddr(&c)
	time.Sleep(time.Second)
	fmt.Println(c)
}

func add(a, b int) (c int) {
	printAddr(&a)
	printAddr(&b)
	printAddr(&c)

	printAddrOff(&b, &c)

	return a + b
}
