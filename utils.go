package jsonschema

type kv[K any, V any] struct {
	k K
	v V
}

// slice 的遍历速度大于map，使用该数据接口来替换map
type sliceMap[K any, V any] struct {
	data []kv[K, V]
}

func (s *sliceMap[K, V]) Len() int {
	return len(s.data)
}

func (s *sliceMap[K, V]) Set(k K, v V) {
	s.data = append(s.data, kv[K, V]{k, v})
}

func (s *sliceMap[K, V]) Range(f func(k K, v V) bool) {
	for _, e := range s.data {
		if !f(e.k, e.v) {
			return
		}
	}
}

func (s *sliceMap[K, V]) Get(k K) (v V, ok bool) {
	for _, e := range s.data {
		if eq(e.k, k) {
			return e.v, true
		}
	}
	return v, false
}

func (s *sliceMap[K, V]) Getv(k K) (v V) {
	v, _ = s.Get(k)
	return
}

func eq(a, b any) bool {
	return a == b
}
