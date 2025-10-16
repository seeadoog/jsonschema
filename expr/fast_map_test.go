package expr

import "testing"

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
