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
	//
}
