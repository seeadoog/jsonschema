package jsonpath

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type index interface {
	get(parent any) (any, bool)
	set(ppk index, pp any, parent any, value any) error
	new() any
}

type indexMap string

func (k indexMap) get(parent any) (any, bool) {
	m, ok := parent.(map[string]interface{})
	if !ok {
		return nil, false
	}
	res, ok := m[string(k)]
	return res, ok
}

func (k indexMap) set(ppk index, pp, parent any, value any) error {
	m, ok := parent.(map[string]interface{})
	if !ok {
		return errors.New("parent is not a map")
	}
	m[string(k)] = value
	return nil
}

func (k indexMap) new() any {
	return make(map[string]any)
}

type indexSlice int

func newIndexSlice(i int) indexSlice {
	return indexSlice(i)
}

func (k indexSlice) get(parent any) (any, bool) {
	m, ok := parent.([]interface{})
	if !ok {
		ps, ok := parent.(*sliceP)
		if ok {
			return ps.get(int(k))
		}
		return nil, false
	}
	if len(m) <= int(k) {
		return nil, false
	}
	if k < 0 {
		if len(m) > 0 {
			return nil, false
			//return m[len(m)-1], true
		}
		//k.i = 0
		return nil, false
	}
	return m[k], true
}

type resetParentError struct {
	parent any
}

func (e *resetParentError) Error() string {
	return "parent should reset"
}

func (k indexSlice) set(ppk index, ppv, parent any, value any) error {
	m, ok := parent.([]interface{})
	if !ok {
		ps, ok := parent.(*sliceP)
		if ok {
			ps.set(int(k), value)
			return nil
		}
		return errors.New("parent is not an array")
	}
	if len(m) <= int(k) {
		m = append(m, make([]interface{}, int(k)-len(m)+1)...)
		//m[k] = value
		if ppk != nil {
			err := ppk.set(nil, nil, ppv, m)
			if err != nil {
				return err
			}
		} else {
			return errors.New("invalid index0")
		}
	} else if k < 0 {
		m = append(m, value)
		if ppk != nil {
			err := ppk.set(nil, nil, ppv, m)
			if err != nil {
				return err
			}
		} else {
			return errors.New("invalid index1")
		}
		return nil

	}
	m[k] = value
	return nil
}

func (k indexSlice) new() any {
	if k < 0 {
		return make([]any, 0)
	}
	return make([]any, k+1)
}

type Complied struct {
	indexes []index
	raw     string
}

func (c *Complied) String() string {
	return c.raw
}

func (c *Complied) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.raw)
}

func (c *Complied) UnmarshalJSON(bytes []byte) error {
	var expr string

	err := json.Unmarshal(bytes, &expr)
	if err != nil {
		return err
	}
	c.raw = expr
	idx, err := compileExpr(expr)
	if err != nil {
		return err
	}
	c.indexes = idx.indexes
	return nil

}

func (c *Complied) Get(src any) (res any, ok bool) {
	for _, idx := range c.indexes {
		src, ok = idx.get(src)
		if !ok {
			return nil, false
		}
	}
	return src, true
}

func Compile(jsonpath string) (*Complied, error) {
	return compileExpr(jsonpath)
}

func (c *Complied) Set2(src any, value any) error {
	if len(c.indexes) == 0 {
		return nil
	}
	index0 := c.indexes[0]
	switch v := src.(type) {
	case *any:
		var o any
		if v == nil || *v == nil {
			o = index0.new()
		} else {
			o = *v
		}
		switch ov := o.(type) {
		case map[string]any:
			err := c.Set(o, value)
			if err != nil {
				return err
			}
			*v = ov
			return nil
		case []any:
			p := &sliceP{data: ov}
			err := c.Set(p, value)
			if err != nil {
				return err
			}
			*v = p.data
			return nil
		}
		return fmt.Errorf("unknown type: %T", o)
	}
	return c.Set(src, value)
}

func (c *Complied) Set(src any, value any) error {
	var parent any
	var ppk index
	for i, idx := range c.indexes {
		if i < len(c.indexes)-1 {
			data, ok := idx.get(src)
			if !ok || data == nil {
				next := c.indexes[i+1]
				data = next.new()
				err := idx.set(ppk, parent, src, data)
				if err != nil {
					return err
				}
			}

			parent = src
			src = data
			ppk = idx

		} else {
			return idx.set(ppk, parent, src, value)
		}
	}
	return nil
}

const (
	scanMap   = 0
	scanSlice = 1
	skip      = 2
)

func compileExpr(expr string) (*Complied, error) {

	token := make([]byte, 0)
	cmp := new(Complied)
	cmp.raw = expr
	status := 0
	for i := 0; i < len(expr); i++ {
		c := expr[i]

		switch status {
		case scanMap:
			if c == '\\' {
				status = skip
				continue
			}
			token = append(token, c)
			if c == '.' {
				tkn := string(token[:len(token)-1])
				if len(tkn) > 0 {
					idx, err := parseToken(tkn, scanMap)
					if err != nil {
						return nil, err
					}
					cmp.indexes = append(cmp.indexes, idx)
				}
				token = token[:0]
			}
			if c == '[' {
				tkn := string(token[:len(token)-1])
				if len(tkn) > 0 {
					idx, err := parseToken(tkn, scanMap)
					if err != nil {
						return nil, err
					}
					cmp.indexes = append(cmp.indexes, idx)
				}
				status = scanSlice
				token = token[len(token)-1:]
			}
		case scanSlice:
			token = append(token, c)
			if c == ']' {
				tkn := string(token[1 : len(token)-1])
				idx, err := parseToken(tkn, scanSlice)
				if err != nil {
					return nil, err
				}
				cmp.indexes = append(cmp.indexes, idx)
				token = token[:0]
				status = scanMap
			}
		case skip:
			token = append(token, c)
			status = scanMap

		}
	}
	if len(token) > 0 {
		idx, err := parseToken(string(token), scanMap)
		if err != nil {
			return nil, err
		}
		cmp.indexes = append(cmp.indexes, idx)
	}
	return cmp, nil
}

func parseToken(token string, status int) (index, error) {
	switch status {
	case scanMap:
		return indexMap(token), nil
	case scanSlice:
		n, err := strconv.Atoi(token)
		if err != nil {
			return nil, err
		}
		return indexSlice(n), nil
	}
	return nil, errors.New("invalid status")
}

type sliceP struct {
	data []any
}

type callf struct {
	st []any // 1 2
}

func (s *sliceP) set(i int, data any) {
	if i < 0 {
		s.data = append(s.data, data)
		return
	}
	if i >= len(s.data) {
		s.data = append(s.data, make([]any, i-len(s.data)+1)...)
	}
	s.data[i] = data
}

func (s *sliceP) get(i int) (any, bool) {
	if i < 0 {
		return nil, false
	}
	if i >= len(s.data) {
		return nil, false
	}
	return s.data[i], true
}

// abc(acsd,'ss',call())
// -- common
