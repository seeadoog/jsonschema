package jsonschema

import (
	"fmt"
	"strings"
)

func init() {
	RegisterValidator("$ref", newRef)
}

type ref struct {
	path   []string
	jp     string
	parent Validator
}

func (r *ref) isSelf(n Validator) bool {
	return n == r || n == r.parent
}

func (r *ref) Validate(c *ValidateCtx, value interface{}) {
	node := c.root
	for _, pth := range r.path {
		switch nv := node.(type) {
		case Children:
			node = nv.GetChild(pth)
		default:
			if r.isSelf(nv) {
				c.AddError(Error{
					Path: r.jp,
					Info: "self reference of $ref",
				})
				return
			}
			node = nil
		}
	}
	if r.isSelf(node) {
		c.AddError(Error{
			Path: r.jp,
			Info: "self reference of $ref",
		})
		return
	}
	cc := c.Clone()
	if node != nil {
		node.Validate(cc, value)
	}
	if len(cc.errors) > 0 {
		for i, e := range cc.errors {
			if len(e.Path) >= 1 {
				p := r.jp + e.Path[1:]
				cc.errors[i] = Error{
					Path: p,
					Info: e.Info,
				}
			}
		}
		c.AddErrors(cc.errors...)
	}

}

var newRef NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {
	str, ok := i.(string)
	if !ok {
		return nil, fmt.Errorf("%s.$ref should be string", path)
	}
	str = strings.TrimPrefix(str, "#")
	str = strings.TrimPrefix(str, "/")
	ref := &ref{
		jp:     path,
		parent: parent,
	}
	if str == "" {
		return ref, nil
	}
	ref.path = strings.Split(str, "/")
	return ref, nil
}
