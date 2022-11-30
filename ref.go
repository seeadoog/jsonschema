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

func (r *ref) Validate(c *ValidateCtx, value interface{}) {
	node := c.root
	for _, pth := range r.path {
		switch nv := node.(type) {
		case Children:
			node = nv.GetChild(pth)
		case *Properties:
			node = nv.properties[pth]
		default:
			if r == nv || r.parent == nv {
				c.AddError(Error{
					Path: r.jp,
					Info: "self reference of $ref",
				})
				return
			}
			node = nil
		}
	}
	if r == node || r.parent == node {
		c.AddError(Error{
			Path: r.jp,
			Info: "self reference of $ref",
		})
		return
	}
	if node != nil {
		node.Validate(c, value)
	}
	if len(c.errors) > 0 {
		for i, e := range c.errors {
			p := r.jp + e.Path[1:]
			c.errors[i] = Error{
				Path: p,
				Info: e.Info,
			}
		}
	}

}

var newRef NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {
	str, ok := i.(string)
	if !ok {
		return nil, fmt.Errorf("%s.$ref should be string", path)
	}
	str = strings.TrimPrefix(str, "#/")
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
