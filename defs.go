package jsonschema

import "fmt"

func init() {
	RegisterValidator("$defs", newDefs)
	RegisterValidator("definitions", newDefs)
}

type Children interface {
	GetChild(path string) Validator
}

type defs struct {
	schemas map[string]Validator
}

func (d *defs) Validate(c *ValidateCtx, value interface{}) {

}

func (d *defs) GetChild(path string) Validator {
	return d.schemas[path]
}

var newDefs NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {
	def := &defs{
		schemas: map[string]Validator{},
	}
	switch v := i.(type) {
	case map[string]interface{}:
		for name, prop := range v {
			vad, err := NewProp(prop, "$")
			if err != nil {
				return nil, fmt.Errorf("create $def err:%w path:%s", err, path)
			}
			def.schemas[name] = vad

		}
	default:
		return nil, fmt.Errorf("$def, type should be object")
	}
	return def, nil
}
