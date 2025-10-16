package expr

type elem struct {
	key Type
	val *funcMap
}

type typeFuncMap struct {
	data [][]elem
	mod  uintptr
}

func newTypeMap(size int) *typeFuncMap {
	return &typeFuncMap{
		data: make([][]elem, size),
		mod:  uintptr(size - 1),
	}
}

func (f *typeFuncMap) put(key Type, val *funcMap) {
	idx := uintptr(key) & f.mod
	for i, e := range f.data[idx] {
		if e.key == key {
			f.data[idx][i].val = val
			return
		}
	}
	f.data[idx] = append(f.data[idx], elem{key: key, val: val})
}

func (f *typeFuncMap) get(key Type) *funcMap {
	idx := uintptr(key) & f.mod
	for _, e := range f.data[idx] {
		if e.key == key {
			return e.val
		}
	}
	return nil
}

func (f *typeFuncMap) foreach(fun func(*funcMap) bool) {
	for _, datum := range f.data {
		for _, e := range datum {
			if !fun(e.val) {
				return
			}
		}
	}
}

type felem struct {
	keyHash uint64
	keyStr  string
	val     *objectFunc
}

type funcMap struct {
	data [][]felem
	mod  uint64
}

func newFuncMap(size int) *funcMap {
	return &funcMap{
		data: make([][]felem, size),
		mod:  uint64(size - 1),
	}
}

func (f *funcMap) put(key uint64, skey string, val *objectFunc) {
	idx := key & f.mod
	for i, e := range f.data[idx] {
		if e.keyHash == key {
			if e.keyStr != skey {
				panic("bug hash conflict :" + skey + ":" + e.keyStr)
			}
			f.data[idx][i].val = val
			return
		}
	}
	f.data[idx] = append(f.data[idx], felem{
		keyHash: key,
		keyStr:  skey,
		val:     val,
	})
}

func (f *funcMap) get(key uint64) *objectFunc {
	idx := key & f.mod
	for _, e := range f.data[idx] {
		if e.keyHash == key {
			return e.val
		}
	}
	return nil
}

func (f *funcMap) foreach(fun func(key string, val *objectFunc) bool) {
	for _, e := range f.data {
		for _, ee := range e {
			if !fun(ee.keyStr, ee.val) {
				return
			}
		}
	}
}

func (f *funcMap) getS(key uint64, ks string) *objectFunc {
	idx := key & f.mod
	for _, e := range f.data[idx] {
		if e.keyStr == ks {
			return e.val
		}
	}
	return nil
}
