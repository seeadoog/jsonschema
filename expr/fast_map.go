package expr

import "fmt"

type elem struct {
	key Type
	val *funcMap
}

type typeFuncMap struct {
	data [][]elem
	mod  uintptr
	size int
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
	f.size++

	if f.size > len(f.data)/8 {
		f.reHash()
	}
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
func (f *typeFuncMap) reHash() {
	old := f.data
	f.data = make([][]elem, len(old)*2)
	f.mod = uintptr(len(old)*2 - 1)
	for _, felems := range old {
		for _, e := range felems {
			f.size--
			f.put(e.key, e.val)
		}
	}
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
	size int
}

func newFuncMap(size int) *funcMap {
	return &funcMap{
		data: make([][]felem, size),
		mod:  uint64(size - 1),
	}
}

func (f *funcMap) reHash() {
	old := f.data
	f.data = make([][]felem, len(old)*2)
	f.mod = uint64(len(old)*2 - 1)
	for _, felems := range old {
		for _, e := range felems {
			f.size--
			f.put(e.keyHash, e.keyStr, e.val)
		}
	}
}

func (f *funcMap) puts(key string, val *objectFunc) {
	f.put(calcHash(key), key, val)

}

func (f *funcMap) put(key uint64, skey string, val *objectFunc) {
	idx := key & f.mod
	for i, e := range f.data[idx] {
		if e.keyHash == key {
			if e.keyStr != skey {
				panic(fmt.Sprintf("hash conflicted '%s' : '%s'  please rename func '%s'", e.keyStr, skey, skey))
			}
			f.data[idx][i].val = val
			return
		}
	}
	f.size++
	f.data[idx] = append(f.data[idx], felem{
		keyHash: key,
		keyStr:  skey,
		val:     val,
	})
	if f.size > len(f.data)/8 {
		f.reHash()
	}
}

func (f *funcMap) get(key uint64) *objectFunc {
	idx := key & f.mod

	//for i := 0; i < len(f.data[idx]); i++ {
	//	e := f.data[idx][i]
	//	if e.keyHash == key {
	//		return e.val
	//	}
	//}
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

type envMapElem struct {
	key     string
	keyHash uint64
	val     any
}

type envMap struct {
	data [][]*envMapElem
	mod  uint64
	size int
}

func newEnvMap(size int) *envMap {

	return &envMap{
		data: make([][]*envMapElem, size),
		mod:  uint64(size - 1),
	}
}

func (f *envMap) getHash(key uint64) any {
	idx := key & f.mod
	for _, e := range f.data[idx] {
		if e.keyHash == key {
			return e.val
		}
	}
	return nil
}

func (f *envMap) putString(key string, val any) {
	f.putHash(calcHash(key), key, val)
}

func (f *envMap) putHash(key uint64, skey string, val any) {
	idx := key & f.mod
	for _, e := range f.data[idx] {
		if e.keyHash == key {
			if e.key != skey {
				panic(fmt.Sprintf("hash conflicted '%s' : '%s'  please rename func '%s'", e.key, skey, skey))
			}
			e.val = val
			return
		}
	}
	f.size++
	f.data[idx] = append(f.data[idx], &envMapElem{
		keyHash: key,
		key:     skey,
		val:     val,
	})
	if f.size > len(f.data)/2 {
		f.reHash()
	}
}

func (f *envMap) putHashOnly(key uint64, skey string, val any) {
	idx := key & f.mod
	for _, e := range f.data[idx] {

		if e.keyHash == key {
			//if e.key != skey {
			//	panic(fmt.Sprintf("hash conflicted '%s' : '%s'  please rename func '%s'", e.key, skey, skey))
			//}
			e.val = val
			return
		}
	}
	f.size++
	f.data[idx] = append(f.data[idx], &envMapElem{
		keyHash: key,
		key:     skey,
		val:     val,
	})
	if f.size > len(f.data)/2 {
		f.reHash()
	}
}

func (f *envMap) reHash() {
	old := f.data
	f.data = make([][]*envMapElem, len(old)*2)
	f.mod = uint64(len(old)*2 - 1)
	for _, felems := range old {
		for _, e := range felems {
			f.size--
			f.putHash(e.keyHash, e.key, e.val)
		}
	}
}

func (f *envMap) foreach(fun func(key uint64, hk string, val any) bool) {
	for _, e := range f.data {
		for _, ee := range e {
			if !fun(ee.keyHash, ee.key, ee.val) {
				return
			}
		}
	}
}

func (f *envMap) del(key uint64) {
	idx := key & f.mod
	//idx := uint64()
	for _, e := range f.data[idx] {
		if e.keyHash == key {
			e.val = nil
			break
		}
	}
}

func (f *envMap) clone() *envMap {
	nm := newEnvMap(int(f.mod) + 1)

	f.foreach(func(key uint64, hk string, val any) bool {
		nm.putHash(key, hk, val)
		return true
	})
	return nm
}
