package expr

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestStr(t *testing.T) {
	e, err := ParseFromJSONStr(`
[
"a = 'hello   world'",
"b = str.fields(a)",
"c = str.join(b,':')",
"d = str.trim(' hello ')",
"e = (' hello ')::trim_left(' ')",
"f = (' hello ')::trim_right(' ')",
"g = str.to_upper('hello')",
"h = str.to_lower('HELLO')",
"i->a->b = 'gg'",
"j.a.b = 'gg'"
]
`)
	if err != nil {
		panic(err)
	}

	c := NewContext(map[string]any{})
	c.ForceType = false
	err = c.Exec(e)
	if err != nil {
		panic(err)
	}

	assertDeepEqual(t, c, "b", []string{"hello", "world"})
	assertDeepEqual(t, c, "c", "hello:world")
	assertDeepEqual(t, c, "d", "hello")
	assertDeepEqual(t, c, "e", "hello ")
	assertDeepEqual(t, c, "f", " hello")
	assertDeepEqual(t, c, "g", "HELLO")
	assertDeepEqual(t, c, "h", "hello")
	assertDeepEqual(t, c, "i.a.b", "gg")
	assertDeepEqual(t, c, "j.a.b", "gg")
}

func TestHttp(t *testing.T) {
	go func() {
		http.HandleFunc("/post", func(writer http.ResponseWriter, request *http.Request) {
			request.ParseForm()

			bs, _ := io.ReadAll(request.Body)
			var i any
			json.Unmarshal(bs, &i)
			bd := map[string]any{
				"h1":   request.Header.Get("h1"),
				"p1":   request.Form.Get("p1"),
				"body": i,
			}
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusOK)
			json.NewEncoder(writer).Encode(bd)
		})
		http.ListenAndServe(":19802", nil)
	}()
	time.Sleep(1 * time.Second)
	e, err := ParseFromJSONStr(`
[
"res = http.request('POST', 'http://127.0.0.1:19802/post?p1=p1',{'h1':'h1'},{name:'xn'},2000)"
]
`)
	if err != nil {
		panic(err)
	}

	c := NewContext(map[string]any{})
	c.ForceType = false
	err = c.Exec(e)
	if err != nil {
		panic(err)
	}
	assertDeepEqual(t, c, "res.json.h1", "h1")
	assertDeepEqual(t, c, "res.json.p1", "p1")
	assertDeepEqual(t, c, "res.json.body.name", "xn")
}

func mapEq(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		bv := b[k]
		switch vv := bv.(type) {
		case map[string]interface{}:
			if !mapEq(b, vv) {
				return false
			}
		default:
			if !reflect.DeepEqual(v, vv) {
				return false
			}
		}
	}
	return true
}

func TestTime(t *testing.T) {
	e, err := ParseFromJSONStr(`
[
"tm = time.parse('2006-01-02 15:04:05','2025-01-02 12:10:20')",
"y = tm::year()",
"m = tm::month()",
"d = tm::day()",
"h = tm::hour()",
"i = tm::minute()",
"s = tm::second()"
]
`)
	if err != nil {
		panic(err)
	}

	c := NewContext(map[string]any{})
	c.ForceType = false
	err = c.Exec(e)
	if err != nil {
		panic(err)
	}

	assertEqual2(t, c.Get("tm") == nil, false)
	assertEqual(t, c, "y", float64(2025))
	assertEqual(t, c, "m", float64(1))
	assertEqual(t, c, "d", float64(2))
	assertEqual(t, c, "h", float64(12))
	assertEqual(t, c, "i", float64(10))
	assertEqual(t, c, "s", float64(20))
}
