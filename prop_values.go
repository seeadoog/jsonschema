package jsonschema

import "fmt"

type defaultVals struct {
	vals map[string]any
}

func (d *defaultVals) Validate(c *ValidateCtx, value interface{}) {
	mm, ok := value.(map[string]interface{})
	if !ok {
		return
	}
	for key, val := range d.vals {
		if _, ok := mm[key]; !ok {
			mm[key] = val
		}
	}
}

var NewDefaultValues NewValidatorFunc = func(i interface{}, path string, parent Validator) (Validator, error) {
	mm, ok := i.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("defaultVals should be object %s", path)
	}

	return &defaultVals{mm}, nil
}
