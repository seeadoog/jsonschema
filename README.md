## Introduction
This is a high-performance jsonschema implementation in Golang, achieving zero memory allocation during validation. It offers a performance boost of more than 10x compared to other open-source versions. Additionally, it supports a rule engine, allowing for the definition of complex validation rules and parameter conversion logic.

## Features

- Supports custom validators.
-	Can generate JSON schemas from Go structs.
-	Zero memory allocation during validation runtime.
-	Allows dynamic changes to JSON values, including setting default values.
-	Supports JSON parsing and setting default values.
-	Supports logical checks and dynamically setting JSON values during validation.
-	Not all standard schema features are fully implemented (no support for reference syntax, and some validators are not implemented).
-	Supports rule engine for dynamic value setting and custom complex logic.

## Benchmark 
This JSONSchema implementation is one of the fastest available. Below is a performance comparison with some mainstream open-source versions,
such as github.com/qri-io/jsonschema and github.com/xeipuuv/gojsonschema.

````
goos: darwin
goarch: amd64
pkg: github.com/seeadoog/jsonschema/v2
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
BenchmarkSchema_local-12                         1236652               955.0 ns/op             0 B/op          0 allocs/op
BenchmarkSchema_gojsonschema-12                    74304             15591 ns/op            7484 B/op        258 allocs/op
BenchmarkSchema_qri_io_jsonschema-12               54739             21301 ns/op           14601 B/op        310 allocs/op
PASS
````

## Installation
````
go get github.com/seeadoog/jsonschema/v2 
````

## QuickStart

```go
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
```

## [License](./LICENSE)

## Supported Validator Examples

#### type 

Specifies the data type of a field 
Valid values: string, number, bool, object, array, integer.

```json
{
  "type": "string"
}
```
or 
```json
{
  "type": "string|number"
}
```

#### properties
Defines the structure of an object. When set to object, the field must conform to the defined properties. If undefined properties should be allowed, you can add "additionalProperties": true.

```json
 {
  "type": "object",
  "properties": {
    "name": {
        "type": "string"
    }
  },
  "additionalProperties": true
}
```

#### maxLength

Specifies the maximum length of a string or array.

#### minLength

Specifies the minimum length of a string or array.

#### maximum

Defines the maximum value for numeric fields.

#### minimum

Defines the minimum value for numeric fields.

#### enum

An array specifying the allowed values.

````json
{
  "enum": ["1","2","3"]
}
````

#### required

An array of strings specifying fields that must be present.
````json
{
  "required": ["username","password"]
}
````

#### pattern

Specifies a regular expression that the value must match.
````json
{
  "type": "string",
  "pattern": "^\\d+$"
}
````

#### multipleOf

Requires the value to be a multiple of the given number.
````json
{
  "type": "number",
  "multipleOf": 5
}
````

#### items

Specifies the validation rules for each item in an array.
```json
{
  "type": "array",
  "items": {
      "type": "object",
      "properties":{
        "username": {
            "type": "string"
        }
      }
  }
}
```

#### switch
A conditional validator. Depending on the value of a key, it applies different validation rules.

```json

{
  "switch": "name",
   "case": {
      "name1": {
        "required": ["age1"]
      } ,
      "name2": {
        "required": ["age2"]
      }

   },
   "defaults": {
      "required": ["key3"]
   }
}

```

#### if

If the if condition passes without errors, the then validator is applied; otherwise, the else validator is used. Errors in if are not raised.

```json
{
  "if": {"required": "key1"},
  "then":{"required": "key2"},
  "else": {"required": "key3"}
}
```

#### dependencies

Specifies dependent fields that must be present when a particular field is set.

```json
{
  "dependencies": {
      "key1": ["key2","key3"]
}
}
```

#### not

The validation passes if the not condition fails.

```json
{
  "not": {
      "type": "string"
  }
}
```

### allOf

Validation passes only if all conditions are met.

```json
{
  "allOf": [
    {
        "type": "string"
    },{
        "maxLength": 50
}
  ]
}
```

### anyOf

Validation passes if any of the conditions are met.
```json
{
  "anyOf": [
    {
        "type": "string"
    },{
        "maxLength": 50
}
  ]
}
```

#### constVal

Parameter conversion validator: the field value is replaced by the value in constVal.

```json
{
    "type": "object",
    "properties": {
         "name":{
              "type": "string",
              "constVal": "alen"
          }
    }
}
```

#### defaultVal

Parameter conversion validator: if the field is missing, it is added with the value from defaultVal.

```json
{
    "type": "object",
    "properties": {
         "name":{
              "type": "string",
              "defaultVal": "alen"
          }
    }
}
```

#### replaceKey

Parameter conversion validator: the value is copied and renamed to the key specified by replaceKey.

```json
{
    "type": "object",
    "properties": {
         "name":{
              "type": "string",
              "replaceKey": "alen"
          }
    }
}
```


#### Custom Logic and JSON Conversion

````json
{
  "type": "object",
  "properties": {
    "name": {
      "type": "string"
    },
    "age": {
      "type": "integer"
    }
  },
  "allOf": [
    {
      "if": {
        "gt": {
          "age": 20
        },
        "lt": {
          "age": 50
        }
      },
      "then": {
        "set": {
          "is_stronger": true
        }
      }
    },
    {
      "if": {
        "gt": {
          "age": 5
        },
        "lt": {
          "age": 15
        }
      },
      "then": {
        "set": {
          "is_child": true
        }
      }
    }
  ]
}
````

#### Other Validators

Refer to the official JSON Schema documentation for additional validators.

### Custom Validators

1.	Implement the Validator interface.
2.	Create a new validator function using NewValidatorFunc.
3.	Register the validator with RegisterValidator(name string, fun NewValidatorFunc).
````go
type Validator interface {
	Validate(c *ValidateCtx, value interface{})
}

type NewValidatorFunc func(i interface{}, path string, parent Validator) (Validator, error)




````

### Generate Schema from Struct
```
type User struct {
    Name   string   `json:"name" maxLength:"15" pattern:"^[0-9a-zA-Z_\\-.]+$"`
    Age    int      `json:"age" maximum:"150" minimum:"1" multipleOf:"2"`
    Childs []string `json:"childs"`
}

sc, err := GenerateSchema(&User{})
if err != nil {
    panic(err)
}

fmt.Println(string(sc.Bytes()))
```

Generated schema:

````json
{
    "properties": {
        "age": {
            "maximum": 150,
            "minimum": 1,
            "multipleOf": 2,
            "type": "integer"
        },
        "childs": {
            "items": {
                "type": "string"
            },
            "type": "array"
        },
        "name": {
            "maxLength": 15,
            "pattern": "^[0-9a-zA-Z_\\-.]+$",
            "type": "string"
            }
        },
    "type": "object"
}

````

### Advanced Validators

```` 
{
    "if":{
        "eq":{
            "username":"root"
        },
        "lt":{
            "age":30
        }
    },
    "then":{
        "error":"root user age should be < 30"
    },
    "and":[
      {
        "if":{
          "neq":{
            "class":"",
            "username":""
          }
        },
        "then":{
          "set":{
              "desc":"${username}(${class})" ,
              "desc_upper":["str.toUpper()","${username}(${class})"]
          }
        }
      },
      {
        "if":{
          "ipIn":{
            "cip":["1.2.3.4"]
          }
        },
        "then":{
          "error":"invalid ip: ${cip}"
        }
      }
    ]
}
````