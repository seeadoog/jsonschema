package expr

import (
	"fmt"
	"strconv"
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

func TestEnvMap(t *testing.T) {
	m := newEnvMap(8)

	for i := 0; i < 10000; i++ {
		ss := strconv.Itoa(i) + "xxxadsf"
		ha := calcHash(ss)
		m.putHash(ha, ss, i)
	}
	confilct := make(map[int][]int)
	for i, datum := range m.data {
		confilct[len(datum)] = append(confilct[len(datum)], i)
	}
	for i, i2 := range confilct {
		if i == 3 {
			fmt.Println(m.data[i2[2]][2].key)
		}
	}

}

func BenchmarkEnvMap(b *testing.B) {
	//5477xxxadsf
	m := newEnvMap(8)

	for i := 0; i < 10000; i++ {
		ss := strconv.Itoa(i) + "xxxadsf"
		ha := calcHash(ss)
		m.putHash(ha, ss, i)
	}
	b.ReportAllocs()

	ha := calcHash("7706xxxadsfsdf")
	m.putHashOnly(ha, "7706xxxadsfsdf", nil)
	for i := 0; i < b.N; i++ {
		m.putHashOnly(ha, "7706xxxadsfsdf", nil)
		//m.getHash(ha)

	}
}
func BenchmarkEnvMap2(b *testing.B) {
	//5477xxxadsf
	var m1 = make(map[string]any, 0)
	for i := 0; i < 10000; i++ {
		ss := strconv.Itoa(i) + "xxxadsf"
		m1[ss] = i
	}
	b.ReportAllocs()

	//ha := calcHash("7706xxxadsf")

	for i := 0; i < b.N; i++ {

		_ = m1["5477xxxadsf"]
	}
	fmt.Println(m1["x"])
}
