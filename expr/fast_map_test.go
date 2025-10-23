package expr

import (
	"fmt"
	"testing"
)

func BenchmarkFS(b *testing.B) {
	f := newFuncMap(4)
	a := "hsd"
	aa := calcHash(a)
	f.put(aa, a, nil)
	for i := 0; i < b.N; i++ {
		//f.getS(aa, a)
		f.get(aa)
	}
}

func TestFF(t *testing.T) {
	return
	for _, datum := range objFuncMap.data {
		if len(datum) > 0 {
			fmt.Println("prt:", len(datum), len(datum[0].val.data), datum[0].val.size)
		}
	}
	objFuncMap.foreach(func(f *funcMap) bool {
		for _, datum := range f.data {
			if len(datum) > 0 {
				fmt.Println(len(datum))
			}
		}
		return true
	})
}

func TestFuncMap(t *testing.T) {
	f := newFuncMap(4)
	f.puts("a", nil)
	f.puts("b", nil)
	f.puts("c", nil)
	f.puts("d", nil)
	f.puts("e", nil)
	f.puts("f", nil)
	f.puts("g", nil)
	f.puts("h", nil)
	f.puts("i", nil)

	assertEqual2(t, f.mod, uint64(127))
	assertEqual2(t, f.size, (9))
}

// redis = new_redis('172.30.89.56:7890,172.223.234.33:8900',const { cluster: true, read_timeout:3000, write_timeout: 3000})
// redis.get('123',const {timeout: 5,}).unwrap()
// redis.set('123',{name: $.name,age: $.age}.to_json_str(),const {timeout: 2000, res_as_json: true}).unwrap(),
// redis.subscribe('')
