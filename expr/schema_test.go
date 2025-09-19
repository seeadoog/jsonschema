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
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(c)

}
