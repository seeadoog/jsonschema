package expr

import (
	"fmt"
	"github.com/cespare/xxhash/v2"
	"hash/crc64"
	"strconv"
	"sync"
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
}

func TestRest(t *testing.T) {
	m := newEnvMap(8)
	m.putString("name", 1)
	m.putString("age", 2)

	assertEqual2(t, m.getString("name"), 1)
	assertEqual2(t, m.getString("age"), 2)

	m.reset()
	assertEqual2(t, m.getString("name"), nil)
	assertEqual2(t, m.getString("age"), nil)
}

func BenchmarkReset(b *testing.B) {
	m := NewContext(nil)
	for i := 0; i < b.N; i++ {
		m.Reset()
	}
}

func BenchmarkReuseVm(b *testing.B) {
	p := sync.Pool{
		New: func() interface{} {
			return NewContext(nil)
		},
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d := p.Get().(*Context)

		d.SetByString("sdfsdf", 1)
		d.SetByString("dsfsdf", 1)

		d.Reset()
		p.Put(d)
	}
}

func BenchmarkParallel(b *testing.B) {
	p := sync.Pool{
		New: func() interface{} {
			return NewContext(nil)
		},
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			d := p.Get().(*Context)
			d.Reset()
			p.Put(d)
		}
	})
}

func TestEnvMap2(t *testing.T) {
	m := newEnvMap(1)
	m.putString("name", 1)
	assertEqual2(t, m.getString("name"), 1)

	m.putString("name", 2)
	assertEqual2(t, m.getString("name"), 2)
	m.putString("age", 3)
	assertEqual2(t, m.getString("age"), 3)
	m.putString("age", 4)
	assertEqual2(t, m.getString("age"), 4)
	m.reset()
	assertEqual2(t, m.getString("name"), nil)
	assertEqual2(t, m.getString("age"), nil)

	m.putString("name", 1)
	assertEqual2(t, m.getString("name"), 1)

	m.putString("name", 2)
	assertEqual2(t, m.getString("name"), 2)
	m.putString("age", 3)
	assertEqual2(t, m.getString("age"), 3)
	m.putString("age", 4)
	assertEqual2(t, m.getString("age"), 4)
	m.putString("age3", 5)
	assertEqual2(t, m.getString("age3"), 5)
	assertEqual2(t, m.size, 3)
	assertEqual2(t, int(m.mod), 7)
}

func TestSyncMap(t *testing.T) {

	m := Map[string, any]{}
	m.Store("name", 1)
	m.Store("age", 2)
	assertEqual2(t, m.Get("name"), 1)
	assertEqual2(t, m.Get("age"), 2)
	m.Store("age", 3)
	assertEqual2(t, m.Get("age"), 3)
	assertEqual2(t, m.Get("age"), 3)
	assertEqual2(t, m.Get("age"), 3)
	m.Delete("name")
	assertEqual2(t, m.Get("name"), nil)

}

// 假设你有这个构造函数
// func newMap() mapp

func TestMapBVT(t *testing.T) {
	m := Map[string, any]{}

	t.Run("Get on empty map", func(t *testing.T) {
		if v := m.Get("not-exist"); v != nil {
			t.Fatalf("expect nil, got %v", v)
		}
	})

	t.Run("Delete on empty map", func(t *testing.T) {
		m.Delete("not-exist")
	})

	t.Run("Basic Store and Get", func(t *testing.T) {
		m.Store("k1", "v1")

		if v := m.Get("k1"); v != "v1" {
			t.Fatalf("expect v1, got %v", v)
		}
	})

	t.Run("Store override existing value", func(t *testing.T) {
		m.Store("k2", 1)
		m.Store("k2", 2)

		if v := m.Get("k2"); v != 2 {
			t.Fatalf("expect 2, got %v", v)
		}
	})

	t.Run("Delete returns old value", func(t *testing.T) {
		m.Store("k3", "old")

		m.Delete("k3")

		if v := m.Get("k3"); v != nil {
			t.Fatalf("expect nil after delete, got %v", v)
		}
	})

	t.Run("Delete twice is safe", func(t *testing.T) {
		m.Store("k4", "v4")

		m.Delete("k4")
	})

	t.Run("Store after Delete", func(t *testing.T) {
		m.Store("k5", "v5")
		m.Delete("k5")
		m.Store("k5", "v5-new")

		if v := m.Get("k5"); v != "v5-new" {
			t.Fatalf("expect v5-new, got %v", v)
		}
	})

	t.Run("Value types", func(t *testing.T) {
		type S struct {
			A int
		}

		s := &S{A: 10}

		m.Store("int", 123)
		m.Store("struct", S{A: 1})
		m.Store("ptr", s)
		m.Store("nil", nil)

		if v := m.Get("int"); v != 123 {
			t.Fatalf("expect 123, got %v", v)
		}

		if v := m.Get("struct"); v.(S).A != 1 {
			t.Fatalf("unexpected struct value: %v", v)
		}

		if v := m.Get("ptr"); v.(*S).A != 10 {
			t.Fatalf("unexpected ptr value: %v", v)
		}

		if v := m.Get("nil"); v != nil {
			t.Fatalf("expect nil value, got %v", v)
		}
	})

	t.Run("Empty string key", func(t *testing.T) {
		m.Store("", "empty-key")

		if v := m.Get(""); v != "empty-key" {
			t.Fatalf("expect empty-key, got %v", v)
		}

		m.Delete("")
	})

	t.Run("Special characters key", func(t *testing.T) {
		key := "中文-key-!@#$%^&*()"
		m.Store(key, "ok")

		if v := m.Get(key); v != "ok" {
			t.Fatalf("expect ok, got %v", v)
		}
	})
}

var (
	_map = map[string]int{}
)

func BenchmarkMap4(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_map = map[string]int{
			"a": 1,
			"b": 1,
			"c": 1,
			"d": 1,
			"e": 1,
		}
	}
}

func BenchmarkHash6(b *testing.B) {
	for i := 0; i < b.N; i++ {

		xxhash.Sum64String("hello world")

	}
}

func BenchmarkHash67(b *testing.B) {
	data := []byte("hello world")
	for i := 0; i < b.N; i++ {
		crc64.Checksum(data, table)
	}
}

func mapLogic(m map[string]any) {
	chanVal, _ := m["chan"].(string)
	funcVal, _ := m["func"].(string)

	if chanVal == "iat" && funcVal == "cbm" {
		appid, _ := m["appid"].(string)
		useOld, _ := m["use_old"].(bool)

		switch appid {
		case "super":
			m["pass"] = true
		case "forbid":
			m["pass"] = false
		case "root":
			m["pass"] = useOld || false
		default:
			m["pass"] = false
		}
	}
}

func Benchmark_MapLogic(b *testing.B) {
	m := map[string]any{
		"chan":    "iat",
		"func":    "cbm",
		"appid":   "root",
		"use_old": true,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mapLogic(m)
	}

}
