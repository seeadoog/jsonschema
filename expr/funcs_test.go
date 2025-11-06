package expr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

func TestStr(t *testing.T) {
	e, err := ParseFromJSONStr(`
[
"a = 'hello   world'",
"b = str_fields(a)",
"c = str_join(b,':')",
"d = str_trim(' hello ')",
"e = (' hello ')::trim_left(' ')",
"f = (' hello ')::trim_right(' ')",
"g = str_to_upper('hello')",
"h = str_to_lower('HELLO')",
"i->a->b = 'gg'",
"j.a.b = 'gg'",
"dd = a != 'hell'",
"ee = ddd or 1",
"qs = a.has_prefix('hello')",
"qs2 = a.has_prefix('ahello')",
"sss = str_builder().write('1','2').write('3').string()",
"uup = 'hello'.to_upper()",
"uupl = uup.to_lower()",
"rpl = uupl.replace('he','HE')",
"emptys = ''.is_empty()",
"emptys2 = hasf.is_empty()",
"emptys3 = ee.is_empty()"
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
	assertDeepEqual(t, c, "dd", true)
	assertDeepEqual(t, c, "ee", float64(1))
	assertDeepEqual(t, c, "qs", true)
	assertDeepEqual(t, c, "qs2", false)
	assertDeepEqual(t, c, "sss", "123")
	assertDeepEqual(t, c, "uup", "HELLO")
	assertDeepEqual(t, c, "uupl", "hello")
	assertDeepEqual(t, c, "rpl", "HEllo")
	assertDeepEqual(t, c, "emptys", true)
	assertDeepEqual(t, c, "emptys2", true)
	assertDeepEqual(t, c, "emptys3", false)
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
"res = http_request('POST', 'http://127.0.0.1:19802/post?p1=p1',{'h1':'h1'},{name:'xn'},2000).body.to_json_obj()"
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
	assertDeepEqual(t, c, "res.h1", "h1")
	assertDeepEqual(t, c, "res.p1", "p1")
	assertDeepEqual(t, c, "res.body.name", "xn")
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
"tm = time_parse('2006-01-02 15:04:05','2025-01-02 12:10:20')",
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

func TestCheckFunc(t *testing.T) {

	checkFunction()
}

var ()

func BenchmarkMaps(b *testing.B) {
	//maps := map[Type]int{
	//	TypeOf(float64(1)): 1,
	//}
	maps := make(map[Type]int, 200)
	maps[TypeOf(float64(1))] = 1
	k := TypeOf(1.1)
	for i := 0; i < b.N; i++ {
		_ = maps[k]
	}
	//println(d)
}

type iface struct {
	t, d uintptr
}

func typf(i any) uintptr {
	p := (*iface)(unsafe.Pointer(&i))
	return p.d
}

func TestName(t *testing.T) {
	os.Setenv("1", "22")
	a := typf(2)
	b := typf(len(os.Getenv("1")))
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(len(os.Getenv("1")))

}

func BenchmarkI(b *testing.B) {
	var a uintptr
	for i := 0; i < b.N; i++ {
		a = typf(i)
	}
	fmt.Println(a)
}

// SELECT table_name FROM information_schema.tables  WHERE table_schema = 'public'  ORDER BY table_name;
func TestSort(t *testing.T) {

	e, err := ParseFromJSONStr(`
[
"data = [1,5,6,3,2,4]",
"data.sort({a,b} => a < b)"
]
`)
	if err != nil {
		panic(err)
	}

	c := NewContext(map[string]any{
		"data2": []int{1, 2, 5, 4, 3},
	})
	c.ForceType = false
	err = c.Exec(e)
	if err != nil {
		//panic(err)
	}

	assertDeepEqual(t, c, "data", []any{1.0, 2.0, 3.0, 4.0, 5.0, 6.0})
}

func TestKK(t *testing.T) {
	return
	dps := objFuncMap

	for _, datum := range dps.data {
		if len(datum) > 0 {
			fmt.Println(len(datum))
		}
		for _, e := range datum {
			for _, felems := range e.val.data {
				if len(felems) > 1 {
					fmt.Println("sub", len(felems))
				}
			}
		}
	}
}

type MyStruct struct {
	D any
}

func TestHash(t *testing.T) {
	var vv *string
	v := &MyStruct{D: vv}
	vvv, err := json.Marshal(v)
	fmt.Println(string(vvv), err)
}
