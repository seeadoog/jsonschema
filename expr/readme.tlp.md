


## Powerful Golang Expr Engine

```
a + b + c + (d*e)
name == '500'? 3 : 4 
aso = 4 
str_has_prefix(name,'123')
arr[0:3]
arr[:3]
arr[2:]
arr[3]
name = 5   # asign 5 to name 
a=4 ; b=5 ; c=6  #执行多个表达式，会返回最后一个表达式的值。
name or 'hello' 
a == 5 && b == 6
a == 5 || b == 6
a != 5 
!ok 
name.hex() # 调用string 的 hex() 方法，返回base16 编码

arr[3] = 5  
object.name = 5  # . 在变量中，会使用jsonpath 来取值，在函数名中则是普通字符，没有其他意义。
object['name'] = 5 # 赋值
object::name = 5  #赋值
object->name = 5  #赋值  :: 和 -> 含义一致

object.name   #取值
object['name'] #取值
object::name #取值
object->name #取值

a=5 #注释，支持注释
'hello world'  # str 定义： 支持三种来包住str
`hello world`
"hello world"

'hello world ${name} time is ${time.now()}' # 变量嵌入字符串

dd = name == 'hello' ? 'abc' : dname or 'cname'  # 三元表达式
arr = [1,2,3,4,5] # 数组定义
obj = { name: '5' ,age: 8 } # map 定义

arr.for( {k,v} => print(k,v)) #lambda 表达式

```

### Usage
````go 
package main

import (
	"fmt"
	expr2 "github.com/seeadoog/jsonschema/v2/expr"
	"time"
)

func main() {
	expr2.RegisterDynamicFunc("get_cur_time", 0)
	expr, err := expr2.ParseValue(`'${name}_${get_cur_time()::format("2006-01-02 15:04:05")}'`)
	if err != nil {
		panic(err)
	}
	ctx := expr2.NewContext(map[string]any{
		"name": "hello",
	})
	ctx.SetFunc("get_cur_time", func(ctx *expr2.Context, args ...expr2.Val) any {
		return time.Now()
	})

	n := expr.Val(ctx)
	fmt.Println("result is:", n)

}

````


### 内置函数支持
```
#对象函数

{{.obj_func}}

# 全局函数
{{.global_func}}

```