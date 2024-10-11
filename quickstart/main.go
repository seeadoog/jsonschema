package main

import (
	"encoding/json"
	"fmt"

	"github.com/seeadoog/jsonschema/v2"
)

func main() {
	schema := `
  {
    "type":"object",
    "properties":{
      "name":{
        "type":"string|number",
        "maxLength":5,
        "minLength":1,
        "maximum":10,
        "minimum":1,
        "enum":["1","2"],
        "replaceKey":"name2",
        "formatVal":"string",
        "format":"phone"
      }
    }
  }
  `
	rootSchema := jsonschema.Schema{}

	err := json.Unmarshal([]byte(schema), &rootSchema)
	if err != nil {
		panic(err)
	}

	js := `
  {
    "name":"1"
  }
  `
	var o interface{}
	err = json.Unmarshal([]byte(js), &o)
	if err != nil {
		panic(err)
	}
	fmt.Println(rootSchema.Validate(o))

}
