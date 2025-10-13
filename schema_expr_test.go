package jsonschema

import (
	"fmt"
	"os"
	"testing"
)

func TestScriptSchema_Validate(t *testing.T) {
	sc := `
{
	"type":"object",
    "defaultVals":{
		"name":"xxxxx",
		"smdts":"chenjian"
	},
	"properties":{
		"name":{
			"type":"string"
		},
		"age":{
			"type":"integer",
			"exclusiveMaximum":15 ,
			"maximum":15,
			"minimum":3,
			"exclusiveMinimum":false
		},
		"child":{
			"properties":{
				"name":{
					"defaultVal":"xx"
				}
			},
			"defaultVal":{}
		}
	},

	"script":[
		"$.sms = $.sms ? $.sms: '500'",
		"if(number(int($.age)) != $.age, return(100,'invalid ')) ",
		"$.hd.rtl = $.name == 'dd' && $.age > 20 ? 'teg' : 'seg' "
	]
}

`
	// a = b = c = 5
	ss, err := NewSchemaFromJSON([]byte(sc))
	if err != nil {
		panic(err)
	}

	c := map[string]any{
		"name": "dd",
		"age":  float64(30),
		"sms":  "23",
	}
	err = ss.Validate(c)

}

func getterOf(a any) (f func(key string) any) {
	switch v := a.(type) {
	case map[string]any:
		return func(key string) any {
			return v[key]
		}
	case map[string]string:
		return func(key string) any {
			return v[key]
		}
	default:
		return nil
	}
}

func errs() (err error) {
	s, err := os.Open("sdfsdf")
	if err != nil {
		return
	}
	s.Name()
	return nil
}
func TestErr(t *testing.T) {

	err := errs()
	fmt.Println(err)
}
