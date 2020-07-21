## jsonschema 

- 支持字段类型转化。

#### type  限定字段类型

取值范围：string  number bool object array
```json
{
  "type": "string"
}
```
或者
```json
{
  "type": "string|number"
}
```

#### properties 
当值为object 时起作用。限定object 中字段的模式，不允许出现properties 中未定义的字段

```json
 {
  "type": "object",
  "properties": {
    "name": {
        "type": "string"
    }
  }
}
```

#### flexProperties

当值为object时起作用。限定object 中字段的模式，允许出现flexProperties 中未定义的字段

#### maxLength

当字段为string 或者array 类型时起作用，限定string的最大长度。（字节数）或者数组的最大长度

#### minLength 

当字段为string 或者array 类型时起作用，限定string的最小长度。（字节数）或者数组的最小长度

#### maximum 

当字段为数字类型时字作用，限定数字的最大值

#### minimum 

当字段为数字类型时起作用，限定数字的最小值

#### enum

该值类型为数组。限定值的枚举范围

````json
{
  "enum": ["1","2","3"]
}
````

#### required

该值类型为字符串数组，限定必须存在数组中声明的字段

````json
{
  "required": ["username","password"]
}
````

#### pattern 

当字段的值为字符串是起作用，pattern 的值是一个正则表达式，会校验字段是否和该正则匹配

````json
{
  "type": "string",
  "pattern": "^\\d+$"
}
````

#### items 

当字段的值为数组时起作用，用于校验数组中的每一个实体是否满足该items 中定义的模式

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
当switch中的key的值满足case 中制定的值时，执行case中对应的校验器。如果都不满足，则执行default中的校验器
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
   "default": {
      "required": ["key3"]
   }
}

```

#### if
 当if 中的校验器没有任何错误时，执行then中的校验器，否则执行else中的校验器。 if中的错误不会抛出
 ```json
{
  "if": {"required": "key1"},
  "then":{"required": "key2"},
  "else": {"required": "key3"}
}
 ```

#### dependencies

当传了某个值时，必须传某些值

```json
{
  "dependencies": {
      "key1": ["key2","key3"]
}
}
```