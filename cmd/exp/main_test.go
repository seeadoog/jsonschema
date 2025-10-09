package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestJSON(t *testing.T) {

	je := json.NewDecoder(bytes.NewReader([]byte(`{} {} {}`)))
	for {
		var i any
		err := je.Decode(&i)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(i)
	}

}

type Tem struct {
	A int
	B int
}

// a=5
// a[:3] = 3
func BenchmarkVal(b *testing.B) {
	tm := &Tem{A: 1, B: 2}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		reflect.ValueOf(tm).Kind()
	}
}
