package jsonschema

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestJpsth(t *testing.T) {

	m := map[string]interface{}{
		//"tkn": []any{},
	}

	must(setJP("name[0][0]", m, 1))
	must(setJP("name[0][1]", m, 3))
	must(setJP("name[1][0]", m, 2))
	must(setJP("name[1][1]", m, 4))

	fmt.Println(m)
	//
	bs, _ := json.MarshalIndent(m, "", "\t")
	fmt.Println(string(bs))

}

func setJP(path string, src, value interface{}) error {
	tk, err := parseJpathCompiled(path)
	if err != nil {
		panic(err)
	}
	return tk.Set(src, value)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func BenchmarkSetJP(b *testing.B) {

	b.ReportAllocs()
	jp, err := parseJpathCompiled("name.age.c.d.e.f")
	must(err)
	src := map[string]interface{}{}
	for i := 0; i < b.N; i++ {
		must(jp.Set(src, &i))
	}

}

func TestParseExpr(t *testing.T) {
	fmt.Println(fb(6, 9))
}

// 5
func f(a, b int) int {
	if b == 0 {
		return a
	}
	return f(b, a%b)
}

func fb(a, b int) int {
	for {
		if b == 0 {
			return a
		}
		tp := b
		b = a % b
		a = tp
	}

}

func BenchmarkGG(b *testing.B) {
	for i := 0; i < b.N; i++ {

		f(33, 22)
	}
}

func BenchmarkGo(b *testing.B) {
	for i := 0; i < b.N; i++ {

		fb(33, 22)
	}
}
