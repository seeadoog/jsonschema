package expr

import (
	"fmt"
	"github.com/seeadoog/jsonschema/v2"
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
	InitSchema()
	// a = b = c = 5
	ss, err := jsonschema.NewSchemaFromJSON([]byte(sc))
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

func BenchmarkContert(b *testing.B) {
	var a any
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		a = structValueToVm(false, &i)
	}
	fmt.Println(a)
}
