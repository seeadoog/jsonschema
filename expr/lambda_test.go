package expr

import (
	"fmt"
	"testing"
)

//
//import (
//	"fmt"
//	"testing"
//)
//
//var (
//	testExpr = `
//$.name = 1
//`
//	ctx = NewContext(map[string]interface{}{
//		"a": "hello",
//		"d": "arch",
//		"mm": map[string]any{
//			//"aaa": "1",
//			//"bbb": "2",
//			"sub": []any{1, 2, 3, 4, 5, 6},
//		},
//	})
//)
//
//func init() {
//	RegisterFunc("test2", func(ctx *Context, args ...Val) any {
//		return nil
//	}, 2)
//	//RegisterFunc("print", func(ctx *Context, args ...Val) any {
//	//	return nil
//	//}, -1)
//
//}
//func TestLambda(t *testing.T) {
//	e, err := ParseValue(testExpr)
//	if err != nil {
//		panic(err)
//	}
//
//	ec, err := compileRootValue(e)
//	if err != nil {
//		panic(err)
//	}
//
//	ctx.SetStackAndEnv(ec, "$", map[string]any{
//		"bb": 1,
//	})
//
//	ctx.InitStackSize(ec.StackSize() * 2)
//
//	fmt.Println(ec.Val(ctx))
//	//fmt.Println(ec.Val(ctx))
//	//fmt.Println(ec.Val(ctx))
//	fmt.Println(ctx.table)
//	fmt.Println(ctx.stack)
//	fmt.Println("stack:", ctx.GetFromStack(ec, "b"))
//
//}
//
//func BenchmarkCompiledValue(b *testing.B) {
//	e, err := ParseValue(testExpr)
//	if err != nil {
//		panic(err)
//	}
//
//	ec, err := compileRootValue(e)
//	if err != nil {
//		panic(err)
//	}
//
//	fmt.Println(ec.Val(ctx))
//	b.ReportAllocs()
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		e.Val(ctx)
//	}
//}
//
//func BenchmarkCompiledValueRaw(b *testing.B) {
//	e, err := ParseValue(testExpr)
//	if err != nil {
//		panic(err)
//	}
//	ctx.NewCallEnv = false
//
//	fmt.Println(e.Val(ctx))
//	b.ReportAllocs()
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		e.Val(ctx)
//	}
//}

func BenchmarkCol(b *testing.B) {
	c := NewContext(map[string]any{})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Clone()
	}
}

type Adder interface {
	Add(a, b int) int
}
type PtrAdder[T any] interface {
	*T // T 的指针类型
	Adder
}

func DoAddGeneric[T any, P PtrAdder[T]](a, b int) int {
	d := (new(T))

	p := P(d)
	return p.Add(a, b)
}

func DoAddGeneric2[T any, P interface {
	*T
	Adder
}](a, b int) int {
	d := (new(T))

	p := P(d)
	return p.Add(a, b)
}

type adder struct {
}

func (*adder) Add(a, b int) int {
	return a + b
}
func TestIN(t *testing.T) {

	DoAddGeneric2[adder](1, 2)
}

//go:noinline
func add(a, b int) Adder {
	return new(adder)
}

func BenchmarkIN(b *testing.B) {
	v := 0
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v = add(i, i+1).Add(i, i)
	}
	fmt.Println(v)
}
