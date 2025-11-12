package expr

import (
	"fmt"
	"github.com/cespare/xxhash/v2"
	"hash/crc64"
	"strconv"
	"testing"
)

func BenchmarkHash(b *testing.B) {
	bs := []byte("hello world ")

	fmt.Println(xxhash.Sum64(bs))
	fmt.Println(xxhash.Sum64(bs))
	for i := 0; i < b.N; i++ {
		xxhash.Sum64(bs)
	}
}

func BenchmarkHash2(b *testing.B) {
	bs := []byte("hello world ")
	for i := 0; i < b.N; i++ {
		crc64.Checksum(bs, table)
	}
}

func BenchmarkHash3(b *testing.B) {

	b.ReportAllocs()

	m := make(map[string]any, 64)
	for i := 0; i < 32; i++ {
		m[strconv.Itoa(i)] = i
	}
	c := NewContext(m)
	for i := 0; i < b.N; i++ {
		c.GetByString("32")
	}
}

func BenchmarkFas(b *testing.B) {
	fm := newFuncMap(64)
	for i := 0; i < 32; i++ {
		fm.puts(strconv.Itoa(i), nil)
	}
	k := calcHash("20")
	for i := 0; i < b.N; i++ {
		fm.get(k)
	}
}
func BenchmarkMap3(b *testing.B) {
	fm := newMap2(256)
	for i := 0; i < 32; i++ {
		fm.put(strconv.Itoa(i), i)
	}

	for i := 0; i < b.N; i++ {
		fm.get("32")
	}
}

type map2Elem struct {
	key     string
	val     any
	keyHash uint64
}

type map2 struct {
	data [][]*map2Elem
	mod  uint64
}

func newMap2(cap uint64) *map2 {
	return &map2{
		data: make([][]*map2Elem, cap),
		mod:  cap - 1,
	}
}

func (m *map2) put(key string, val any) {
	idx := xxhash.Sum64(ToBytes(key)) & m.mod

	for _, v := range m.data[idx] {
		if v.key == key {
			v.val = val
			return
		}
	}
	m.data[idx] = append(m.data[idx], &map2Elem{
		key:     key,
		val:     val,
		keyHash: calcHash(key),
	})
}

func (m *map2) get(key string) any {
	idx := xxhash.Sum64(ToBytes(key)) & m.mod

	for _, v := range m.data[idx] {
		if v.key == key {
			return v.val
		}
	}
	return nil
}

func Test222(t *testing.T) {
	fmt.Println(string(byte('.')))
}
