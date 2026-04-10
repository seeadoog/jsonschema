# expr 规则引擎文档

## 目录

- [一、开发者 API](#一开发者-api)
  - [核心类型](#核心类型)
  - [解析函数](#解析函数)
  - [Context 生命周期](#context-生命周期)
  - [错误处理](#错误处理)
  - [注册自定义全局函数](#注册自定义全局函数)
  - [注册自定义对象方法](#注册自定义对象方法)
  - [自定义结构体集成](#自定义结构体集成)
  - [Context 复用（sync.Pool）](#context-复用syncpool)
  - [脚本语法速查](#脚本语法速查)
- [二、内置函数参考](#二内置函数参考)
  - [1. 全局函数](#1-全局函数)
  - [2. 对象函数](#2-对象函数)

---

## 一、开发者 API

### 核心类型

```go
// Expr 是编译后可执行的脚本/表达式，由 Parse* 函数返回
// 编译一次，可多次执行（线程安全地复用）
type Expr  // interface{ Exec(c *Context) error }

// Val 是一个惰性求值节点，调用 Val(c) 才真正求值
type Val interface {
    Val(c *Context) any
}

// ScriptFunc 是所有内置和自定义函数的签名
type ScriptFunc func(ctx *Context, args ...Val) any
```

---

### 解析函数

```go
// 解析 JSON 字符串（数组或字符串）为可执行 Expr
// 数组形式：每个元素为一条语句，按顺序执行
// 字符串形式：单条语句
func ParseFromJSONStr(str string) (Expr, error)

// 解析已反序列化的 Go 值（map/slice/string）为 Expr
func ParseFromJSONObj(o any) (Expr, error)

// 解析单个表达式字符串，返回 Val（可求值节点）
var ParseValue = parseValueV  // func(expr string) (Val, error)

// 解析单个表达式字符串，返回 Expr（可执行节点）
var ParseExpr = parseExpr     // func(expr string) (Expr, error)
```

**示例：**

```go
// JSON 数组脚本（推荐方式）
e, err := expr.ParseFromJSONStr(`[
    "name = 'hello'",
    "age  = 5",
    "result = age > 3 ? 'adult' : 'child'"
]`)

// 单条表达式
e, err := expr.ParseFromJSONStr(`"age > 18"`)

// JSON 对象（if/for/switch 控制流）
e, err := expr.ParseFromJSONObj(map[string]any{
    "if":   "age > 18",
    "then": "label = 'adult'",
    "else": "label = 'child'",
})
```

---

### Context 生命周期

`Context` 是脚本的执行环境，持有变量表。

```go
// 创建 Context，table 是初始变量（可为 nil）
func NewContext(table map[string]any) *Context

// 执行编译好的 Expr（出错时 panic，受 PanicWhenError 控制）
func (c *Context) Exec(e Expr) error

// 安全执行（内部 recover，将 panic 转为 error 返回）
func (c *Context) SafeExec(e Expr) (err error)

// 读取变量
func (c *Context) GetByString(key string) interface{}
func (c *Context) GetByJp(key string) any  // 支持 jsonpath："a.b.c"、"a[0]"

// 写入变量
func (c *Context) SetByString(skey string, value interface{})

// 删除变量
func (c *Context) Delete(key string)

// 获取 return() 的返回值
func (c *Context) GetReturn() []any

// 获取完整变量表
func (c *Context) GetTable() map[string]any

// 重置（清空变量，保留内存分配，用于 sync.Pool 复用）
func (c *Context) Reset()

// 浅克隆（用于并发场景）
func (c *Context) Clone() *Context

// 传入 Go 的 context.Context（用于超时控制）
func (c *Context) SetContext(ctx context.Context)
```

**Context 配置选项：**

```go
c := expr.NewContext(nil)
c.ForceType = false  // false: 结构体字段保持原始类型；true: 强制转换为 VM 类型（float64 等）
c.NewCallEnv = false // true: lambda 调用时使用独立环境（隔离性更强，性能略低）
c.Debug    = false   // true: 输出调试信息
c.IgnoreFuncNotFoundError = false // true: 函数未找到时不报错，返回 nil
```

**完整使用示例：**

```go
package main

import (
    "fmt"
    "github.com/yourorg/jsonschema/expr"
)

func main() {
    // 1. 解析脚本（一次）
    e, err := expr.ParseFromJSONStr(`[
        "score = input * 2",
        "level = score >= 90 ? 'A' : (score >= 60 ? 'B' : 'C')"
    ]`)
    if err != nil {
        panic(err)
    }

    // 2. 创建 Context 并执行（多次复用脚本）
    c := expr.NewContext(map[string]any{
        "input": 50.0,
    })
    if err := c.Exec(e); err != nil {
        panic(err)
    }

    fmt.Println(c.GetByString("score")) // 100
    fmt.Println(c.GetByString("level")) // A
}
```

---

### 错误处理

```go
// PanicWhenError 全局开关（默认 true）
// true: 错误时 panic；false: Exec 返回 error
var PanicWhenError = true

// 脚本运行时错误
type RuntimeError struct{ Err string }
func (r *RuntimeError) Error() string

// 脚本主动返回的错误值（由表达式产生）
type Error struct{ Err any }
func (e *Error) Error() string

// return() 内置函数产生的返回值（通过 GetReturn() 读取）
type Return struct{ Var []any }

// 解析 return() 的返回值
func ValueOfReturn(e error) []any

// 函数返回值封装（Err != nil 表示失败）
type Result struct {
    Err  any `json:"err"`
    Data any `json:"data"`
}
```

---

### 注册自定义全局函数

```go
// 基础注册（argsNum=-1 表示可变参数）
func RegisterFunc(funName string, f ScriptFunc, argsNum int, opts ...funcOpt)

// 类型安全的泛型包装（推荐）
func FuncDefine[R any](f func() R) ScriptFunc
func FuncDefine1[A1, R any](f func(a A1) R) ScriptFunc
func FuncDefine2[A1, A2, R any](f func(a A1, b A2) R) ScriptFunc
func FuncDefine3[A1, A2, A3, R any](f func(a A1, b A2, c A3) R) ScriptFunc
func FuncDefine4[A1, A2, A3, A4, R any](f func(a A1, b A2, c A3, d A4) R) ScriptFunc
func FuncDefine5[A1, A2, A3, A4, A5, R any](f func(a A1, b A2, c A3, d A4, e A5) R) ScriptFunc

// 带 Context 参数的版本
func FuncDefine1WithCtx[A1, R any](f func(ctx *Context, a A1) R) ScriptFunc
func FuncDefine2WithCtx[A1, A2, R any](f func(ctx *Context, a A1, b A2) R) ScriptFunc
// ...以此类推

// 可变参数版本
func FuncDefineN[T, R any](f func(a ...T) R) ScriptFunc

// 带可选 Options 参数（最后一个参数为 map[string]any）
func RegisterOptFuncDefine0[R any](fname string, f func(ctx *Context, opt *Options) R, opts ...commonFuncOpt)
func RegisterOptFuncDefine1[A, R any](fname string, f func(ctx *Context, a A, opt *Options) R, opts ...commonFuncOpt)
func RegisterOptFuncDefine2[A, B, R any](fname string, f func(ctx *Context, a A, b B, opt *Options) R, opts ...commonFuncOpt)
func RegisterOptFuncDefine3[A, B, C, R any](fname string, f func(ctx *Context, a A, b B, c C, opt *Options) R, opts ...commonFuncOpt)
func RegisterOptFuncDefine4[A, B, C, D, R any](fname string, f func(ctx *Context, a A, b B, c C, d D, opt *Options) R, opts ...commonFuncOpt)
```

**注册选项：**

```go
func WithArgsString(s string) funcOpt      // 参数描述字符串（文档用）
func Doc(doc string) commonFuncOpt         // 函数说明（show_doc() 显示）
func WithCompiledArgs[T any](n int, fun func(args ...any) T) commonFuncOpt // 编译期常量折叠
```

**示例：**

```go
// 注册无参函数
expr.RegisterFunc("my_uuid", expr.FuncDefine(func() string {
    return uuid.New().String()
}), 0)

// 注册 2 参函数
expr.RegisterFunc("my_add", expr.FuncDefine2(func(a, b float64) float64 {
    return a + b
}), 2, expr.WithArgsString("(a float64, b float64) float64"))

// 带 Options 的函数（脚本调用：my_func('hello', {timeout: 3000})）
expr.RegisterOptFuncDefine1("my_func", func(ctx *expr.Context, s string, opt *expr.Options) string {
    timeout := opt.GetNumberDef("timeout", 5000)
    _ = timeout
    return s + "_done"
}, expr.Doc("my_func(s string, opt?) string"))
```

---

### 注册自定义对象方法

对象方法通过 `value::method()` 或 `value.method()` 调用。

```go
// 基础注册
func RegisterObjFunc[T any](name string, fun SelfFunc, argsNum int, doc string)

type SelfFunc func(ctx *Context, self any, args ...Val) any

// 类型安全的泛型包装（推荐）
func SelfDefine0[S, R any](name string, f func(ctx *Context, self S) R, opt ...selfDefineOptFunc)
func SelfDefine1[A, S, R any](name string, f func(ctx *Context, self S, a A) R, opt ...selfDefineOptFunc)
func SelfDefine2[A, B, S, R any](name string, f func(ctx *Context, self S, a A, b B) R, opt ...selfDefineOptFunc)
func SelfDefine3[A, B, C, S, R any](name string, f func(ctx *Context, self S, a A, b B, c C) R, opt ...selfDefineOptFunc)
func SelfDefine4[A, B, C, D, S, R any](name string, f func(ctx *Context, self S, a A, b B, c C, d D) R, opt ...selfDefineOptFunc)
func SelfDefineN[S, R any](name string, f SelfFunc)                        // 可变参数

// 为所有类型注册通用方法（慎用，不能与对象方法重名）
func SetFuncForAllTypes(fun string)

// 方法文档选项
func WithDoc(doc string) selfDefineOptFunc
```

**示例：**

```go
type MyObj struct{ Value string }

// 注册 MyObj 的 upper() 方法：myobj.upper()
expr.SelfDefine0("upper", func(ctx *expr.Context, self *MyObj) string {
    return strings.ToUpper(self.Value)
})

// 注册 MyObj 的 prefix(s) 方法：myobj.prefix('hello_')
expr.SelfDefine1("prefix", func(ctx *expr.Context, self *MyObj, s string) string {
    return s + self.Value
}, expr.WithDoc("prefix(string)string  add prefix to value"))
```

---

### 自定义结构体集成

实现以下接口，结构体的字段可在脚本中直接读写（`struct.Name`、`struct.Name = 'new'`）：

```go
// 实现 GetField 和 SetField 接口
type MyData struct {
    Name string
    Age  int
}

func (d *MyData) GetField(ctx *expr.Context, key string) any {
    switch key {
    case "name": return d.Name
    case "age":  return float64(d.Age)
    }
    return nil
}

func (d *MyData) SetField(ctx *expr.Context, name string, val any) {
    switch name {
    case "name": d.Name = expr.StringOf(val)
    case "age":  d.Age = int(expr.NumberOf(val))
    }
}

// 在脚本中使用
c := expr.NewContext(map[string]any{
    "data": &MyData{Name: "alice", Age: 30},
})
// 脚本：data.name   data.age   data.name = 'bob'
```

对于普通 Go 结构体（未实现上述接口），引擎通过反射访问导出字段，使用 `->` 操作符：

```
"a = usr->Name"
"usr->Age = 25"
"a = usr->Friends[0]->Name"
```

---

### Context 复用（sync.Pool）

```go
var pool = sync.Pool{
    New: func() interface{} { return expr.NewContext(nil) },
}

func eval(e expr.Expr, input map[string]any) map[string]any {
    c := pool.Get().(*expr.Context)
    defer func() {
        c.Reset()     // 清空变量，保留内存分配
        pool.Put(c)
    }()

    for k, v := range input {
        c.SetByString(k, v)
    }
    c.Exec(e)
    return c.GetTable()
}
```

---

### 脚本语法速查

#### 变量赋值

```
name = 'hello'           # 字符串
age  = 25                # 数字（float64）
flag = true              # 布尔
obj.key = 'val'          # 嵌套 map 自动创建
arr[0] = 'x'             # 数组索引赋值
a.b.c = 1                # 深路径自动创建
```

#### 运算符

```
+  -  *  /  ^(幂)  %(取模)
&(按位与)  |(按位或)
==  !=  >  >=  <  <=
&&  ||  !
??  (nil 合并：a ?? b，a 为 nil 时返回 b)
a++  a--  a += b
```

#### 三元与条件

```
result = age >= 18 ? 'adult' : 'child'
result = a > 0 ? 'pos' : (a < 0 ? 'neg' : 'zero')
```

#### if / else（JSON 对象写法）

```json
{"if": "score >= 90", "then": "grade='A'", "else": "grade='B'"}
```

#### if / else（内联方法链写法）

```
if(score>=90).then(grade='A').elseif(score>=60,grade='B').else(grade='C').end()
```

#### switch

```json
{"switch": "status", "case": {"'ok'": "flag=1", "'err'": "flag=0"}, "default": "flag=-1"}
```

```
switch(status).case('ok',flag=1).case('err',flag=0).default(flag=-1).end()
```

#### for 循环

```json
{"for": "k,v in items", "do": ["result[k] = v * 2"]}
```

```
for(items, {k,v} => result[k] = v * 2)
```

#### Lambda

```
add = {a, b} => a + b
result = add(3, 4)

# 多行 lambda（用括号包裹多条语句）
process = {x} => (
    tmp = x * 2;
    tmp + 1
)
```

#### 字符串插值

```
msg = '${name} is ${age} years old'
date = '${now::year()}-${now::month()}-${now::day()}'
```

#### 方法调用（`::` 和 `.` 等价）

```
'hello world'::has_prefix('hello')     # true
name::to_upper()                       # 大写
arr::slice(0, 3)::join(',')            # 切片后拼接
time_now()::format('2006-01-02')       # 时间格式化
```

#### 多语句（分号分隔）

```
a=1; b=2; c=a+b
```

#### as 别名

```
5 + 5 as total
name::to_upper() as upper_name
```

#### return 提前退出

```
return('ok')    # 执行结束，GetReturn() 可取到 ["ok"]
```

#### or 空值合并链

```
result = value or default_val or 'fallback'  # 返回第一个真值
```

---

## 二、内置函数参考

> 调用方式：`funcname(arg1, arg2, ...)`  
> 返回类型 `number` 均为 `float64`。

---

### 1. 全局函数

#### 时间函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `time_now()` | `time.Time` | 获取当前时间 |
| `time_now_mill()` | `number` | 获取当前时间的毫秒级 Unix 时间戳 |
| `time_from_unix(unix number)` | `time.Time` | 将秒级 Unix 时间戳转为 `time.Time` |
| `time_format(t time.Time, layout string)` | `string` | 格式化时间，layout 使用 Go 时间格式（如 `"2006-01-02 15:04:05"`） |
| `time_parse(layout string, str string)` | `time.Time` | 解析时间字符串为 `time.Time` |

**使用示例：**
```
now = time_now()
ts  = time_now_mill()
t   = time_from_unix(1700000000)
str = time_format(now, '2006-01-02')
t2  = time_parse('2006-01-02', '2024-01-15')
```

---

#### 类型转换函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `string(v any)` | `string` | 转为字符串 |
| `number(v any)` | `number` | 转为 float64 |
| `int(v any)` | `number` | 转为整数（截断小数） |
| `bool(v any)` / `boolean(v any)` | `bool` | 转为布尔值 |
| `bytes(v any)` | `[]byte` | 转为字节切片 |
| `type(v any)` | `string` | 返回值类型名："string"/"number"/"boolean"/"nil"/"array"/"map" 等 |

**使用示例：**
```
n   = number('3.14')    # 3.14
s   = string(100)       # "100"
b   = bool(0)           # false
t   = type('hello')     # "string"
```

---

#### 数学函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `add(a, b, ...)` | `number` / `string` | 加法或字符串拼接（可变参数） |
| `sub(a number, b number)` | `number` | 减法 |
| `mul(a number, b number)` | `number` | 乘法 |
| `div(a number, b number)` | `number` | 除法 |
| `mod(a number, b number)` | `number` | 取模 |
| `pow(a number, b number)` | `number` | 幂运算（`a^b`） |
| `neg(a number)` | `number` | 取反 |
| `sqrt(n number)` | `number` | 平方根 |
| `log10(n number)` | `number` | 以 10 为底的对数 |

**使用示例：**
```
r1 = pow(2, 10)     # 1024
r2 = sqrt(16)       # 4
r3 = mod(10, 3)     # 1
```

---

#### 比较函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `eq(a, b)` | `bool` | 严格相等（类型和值都相等） |
| `neq(a, b)` | `bool` | 不等 |
| `eqs(a, b)` | `bool` | 转字符串后相等 |
| `neqs(a, b)` | `bool` | 转字符串后不等 |
| `lt(a, b)` | `bool` | 小于 |
| `lte(a, b)` | `bool` | 小于等于 |
| `gt(a, b)` | `bool` | 大于 |
| `gte(a, b)` | `bool` | 大于等于 |
| `inn(val, a, b, ...)` | `bool` | 判断 val 是否在后续参数中（in 操作符） |

**使用示例：**
```
ok = eq(status, 'active')
ok = inn(role, 'admin', 'editor', 'viewer')
```

---

#### 逻辑函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `not(a any)` | `bool` | 逻辑非 |
| `or(a, b, ...)` | `any` | 返回第一个真值（类似 `a \|\| b`） |
| `orr(a, b)` | `any` | 返回第一个非 nil 值 |
| `and(a, b, ...)` | `bool` | 所有参数均为真时返回 true |
| `if(a, b, ...)` | `any` | 所有参数均为真时返回最后一个参数值 |
| `ternary(cond, a, b)` | `any` | 三元：cond 为真返回 a，否则返回 b |

**使用示例：**
```
val  = or(a, b, 'default')
val  = ternary(age > 18, 'adult', 'minor')
safe = orr(user_input, '')
```

---

#### 字符串函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `str_has_prefix(s string, prefix string)` | `bool` | 判断前缀 |
| `str_has_suffix(s string, suffix string)` | `bool` | 判断后缀 |
| `str_join(arr []any, sep string)` | `string` | 数组元素用 sep 拼接为字符串 |
| `str_split(s string, sep string, n number)` | `[]any` | 分割字符串，n 为最大段数（-1 不限） |
| `str_to_upper(s string)` | `string` | 转大写 |
| `str_to_lower(s string)` | `string` | 转小写 |
| `str_trim(s string)` | `string` | 去除首尾空格 |
| `str_fields(s string)` | `[]any` | 按空白字符分割字符串 |
| `sprintf(format string, args ...)` | `string` | 格式化字符串（同 `fmt.Sprintf`） |
| `print(args ...)` | `nil` | 打印到标准输出 |
| `printf(format string, args ...)` | `nil` | 格式化打印到标准输出 |
| `str_builder()` | `*strings.Builder` | 创建字符串构建器 |

**使用示例：**
```
parts = str_split('a,b,c', ',', -1)    # ["a","b","c"]
msg   = sprintf('hello %s, age=%d', name, age)
sb    = str_builder()::write('a')::write('b')::string()  # "ab"
```

---

#### JSON 函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `to_json_str(v any)` / `json_to(v any)` | `string` | 将值序列化为 JSON 字符串 |
| `to_json_obj(s string)` / `json_from(s string)` | `any` | 将 JSON 字符串反序列化为 map/slice |

**使用示例：**
```
s   = to_json_str({'name':'alice','age':30})   # '{"age":30,"name":"alice"}'
obj = to_json_obj('{"key":"val"}')             # map[string]any
```

---

#### 数组函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `append(arr []any, items ...)` | `[]any` | 追加元素到数组（或拼接字符串） |
| `join(arr []any, sep string)` | `string` | 数组拼接为字符串 |
| `len(v any)` | `number` | 返回字符串/数组/map/字节切片的长度 |
| `slice_new(size number)` | `[]any` | 创建指定长度的空数组 |
| `slice_init(items ...)` | `[]any` | 用参数列表初始化数组 |
| `slice_cut(arr []any, start number, end number)` | `[]any` | 切片 `[start:end]` |
| `range(n number)` | `[]any` | 生成 `[0, 1, ..., n-1]` 的数组 |
| `all(arr []any, cond lambda)` | `[]any` | 过滤满足条件的元素 |

**使用示例：**
```
arr = slice_init(1, 2, 3, 4, 5)
sub = slice_cut(arr, 1, 3)          # [2, 3]
idx = range(5)                      # [0,1,2,3,4]
evens = all(arr, {v} => v % 2 == 0) # [2, 4]
```

---

#### Map 函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `new()` | `map[string]any` | 创建新的空 map |
| `get(obj map, key string)` | `any` | 获取 map 的键值 |
| `set(obj map, key string, val any)` | `map[string]any` | 设置 map 键值 |
| `delete(obj map, key string)` | `map[string]any` | 删除 map 的键 |
| `set_index(arr []any, i number, val any)` | `[]any` | 设置数组指定索引的值 |

---

#### 控制流函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `for(collection any, lambda)` | `any` | 遍历数组或 map，lambda 接收 `{k, v}` |
| `loop(cond lambda, do lambda)` / `loop(do lambda)` | `any` | 循环执行，直到 cond 返回 false |
| `return(vals ...)` | — | 立即终止脚本，返回值通过 `GetReturn()` 获取 |
| `repeat(lambda, n number)` | `[]any` | 执行 lambda n 次，收集结果 |
| `repeats(lambda, n number)` | `nil` | 执行 lambda n 次，忽略结果 |

**使用示例：**
```
for(items, {k, v} => total += v)
loop({i} => i < 10, {i} => i++)
result = repeat({} => rand(100), 5)   # 执行 5 次，返回结果数组
```

---

#### 加密与编码函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `base64_encode(v any)` | `string` | Base64 编码（字符串或字节切片） |
| `base64_decode(s string)` | `[]byte` | Base64 解码，失败时 panic |
| `md5_sum(s string)` | `[]byte` | 计算 MD5，返回字节切片 |
| `sha256_sum(s string)` | `[]byte` | 计算 SHA-256，返回字节切片 |
| `hmac_sha256(data any, secret any)` | `[]byte` | 计算 HMAC-SHA256 |
| `hex_encode(b []byte)` | `string` | 字节切片转十六进制字符串 |
| `hex_decode(s string)` | `[]byte` | 十六进制字符串转字节切片，失败时 panic |

**使用示例：**
```
sig = hmac_sha256('data', 'secret')::hex()
h   = md5_sum('hello')::hex()
enc = base64_encode('hello world')
```

---

#### 并发函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `go(lambda)` | `*goroutine` | 在新 goroutine 中执行 lambda，返回 goroutine 句柄 |
| `sleep(ms number)` | `nil` | 休眠指定毫秒数 |
| `defer(do lambda, defer_expr lambda)` | `any` | 执行 do，结束后（无论成功失败）执行 defer_expr |

**使用示例：**
```
g = go({} => expensive_calc())
sleep(100)
result = g::join()        # 等待 goroutine 完成并获取结果
```

---

#### 错误处理函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `catch(v any)` | `any` | 若 v 是 `*Error` 或 `*Return`，返回 nil；否则返回 v |
| `unwrap(v any)` | `any` | 若 v 是 `*Result{Err:...}`，则 panic；否则返回 data |
| `recover(lambda)` | `*Result` | 执行 lambda，捕获 panic，返回 `*Result{Err, Data}` |
| `recovers(lambda)` | `*Result` | 同 recover，但 Err 包含 stack trace |
| `recoverd(lambda)` | `*Result` | 同 recover，但非 nil 结果包装为 `*Result{Data}` |

**使用示例：**
```
res = recover({} => risky_operation())
res::unwrap()         # 有错误时 panic，否则返回 data

safe = catch(may_fail())
```

---

#### HTTP 函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `http_request(method string, url string, headers map, body any, timeout_ms number)` | `map[string]any` | 发送 HTTP 请求，返回 `{body:[]byte, header:map, status:number}`，非 2xx 时 panic |
| `curl(url string, opt map)` | `*httpResp` | 发送 HTTP 请求，返回 `*httpResp`，非 2xx 时 panic |

`curl` 的 opt 选项：

| 键 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `method` | string | GET（无 body）/ POST（有 body） | HTTP 方法 |
| `header` | map | `{}` | 请求头 |
| `body` | any | nil | 请求体（map 自动序列化为 JSON） |
| `ip` | string | "" | 强制使用指定 IP（绕过 DNS） |
| `timeout` | number | 60000 | 超时毫秒数 |
| `ssl_verify` | bool | true | 是否验证 SSL 证书 |

**使用示例：**
```
# 简单 GET
resp = curl('https://api.example.com/users')
data = resp.body::string()::to_json_obj()

# POST JSON
resp = curl('https://api.example.com/create', {
    method: 'POST',
    header: {'Content-Type': 'application/json'},
    body:   {name: 'alice', age: 30},
    timeout: 5000
})
resp::throw()   # 非 2xx 时 panic
```

---

#### 其他实用函数

| 函数签名 | 返回类型 | 说明 |
|---|---|---|
| `is_empty(v any)` | `bool` | 判断 nil、空字符串、空数组是否为空 |
| `rand(n number)` | `number` | 生成 `[0, n)` 的随机整数 |
| `exec(cmd string, args ...)` | `string` | 执行 Shell 命令，返回标准输出 |
| `cost(lambda)` | `number` | 返回 lambda 执行耗时（毫秒） |
| `set_to(val any, varName string)` / `seto` | `any` | 将 val 赋给名为 varName 的变量，同时返回 val |
| `show_doc()` | `nil` | 打印所有已注册的函数和方法文档 |
| `show_env()` | `nil` | 打印当前 Context 的所有变量 |
| `regexp_new(pattern string)` | `*regexp.Regexp` | 编译正则表达式 |
| `url_new_values()` | `url.Values` | 创建 URL 查询参数对象 |
| `atomic_int(n number)` | `*atomic.Int64` | 创建原子整数（并发安全） |

**使用示例：**
```
ms = cost({} => sleep(100))        # ≈ 100

# set_to 可用于在链式调用中保存中间值
result = expensive()::set_to('tmp') + 1
```

---

### 2. 对象函数

> 调用方式：`value::method(args)` 或 `value.method(args)`

---

#### string 字符串方法

| 方法签名 | 返回类型 | 说明 |
|---|---|---|
| `string::has_prefix(prefix string)` | `bool` | 判断是否以 prefix 开头 |
| `string::has_suffix(suffix string)` | `bool` | 判断是否以 suffix 结尾 |
| `string::has(sub string)` | `bool` | 判断是否包含子串 |
| `string::contains(sub string)` | `bool` | 同 has |
| `string::trim_space()` | `string` | 去除首尾空白字符 |
| `string::trim(cutset string)` | `string` | 去除首尾指定字符集 |
| `string::trim_left(cutset string)` | `string` | 去除左侧指定字符集 |
| `string::trim_right(cutset string)` | `string` | 去除右侧指定字符集 |
| `string::trim_prefix(prefix string)` | `string` | 去除前缀 |
| `string::trim_suffix(suffix string)` | `string` | 去除后缀 |
| `string::slice(a number, b number)` | `string` | 子串 `[a:b]` |
| `string::len()` | `number` | 字符串字节长度 |
| `string::split(sep string, n number)` | `[]any` | 分割字符串，n=-1 不限段数 |
| `string::to_upper()` | `string` | 转大写 |
| `string::to_lower()` | `string` | 转小写 |
| `string::replace(old string, new string)` | `string` | 替换所有匹配（`strings.Replace` all） |
| `string::index(sub string)` | `number` | 返回子串首次出现的索引，-1 表示未找到 |
| `string::bytes()` | `[]byte` | 转为字节切片 |
| `string::md5()` | `[]byte` | 计算 MD5 |
| `string::hex()` | `string` | 字符串内容转十六进制 |
| `string::base64()` | `string` | Base64 编码 |
| `string::base64d()` | `[]byte` | Base64 解码 |
| `string::fields()` | `[]string` | 按空白字符分割 |
| `string::json_str()` | `string` | JSON 编码（含引号） |

**使用示例：**
```
'  hello  '::trim_space()          # "hello"
'hello world'::split(' ', -1)      # ["hello","world"]
'hello'::slice(1, 3)               # "el"
'Hello World'::to_lower()          # "hello world"
'abc'::has_prefix('ab')            # true
'hello'::base64()::base64d()::string()  # "hello"
```

---

#### []any 数组方法

| 方法签名 | 返回类型 | 说明 |
|---|---|---|
| `[]any::slice(a number, b number)` | `[]any` | 子切片 `[a:b]` |
| `[]any::len()` | `number` | 数组长度 |
| `[]any::get(i number)` | `any` | 获取索引 i 的元素 |
| `[]any::clone()` | `[]any` | 浅克隆 |
| `[]any::join(sep string)` | `string` | 元素转字符串后用 sep 拼接 |
| `[]any::all(cond lambda)` | `[]any` | 过滤满足条件的元素 |
| `[]any::filter(mapper lambda)` | `[]any` | 将每个元素映射为新值 |
| `[]any::sort(cmp lambda)` | `[]any` | 原地排序，lambda 接收 `{a,b}`，返回 `a < b` 则 a 排前 |
| `[]any::for(lambda)` | `any` | 遍历，lambda 接收 `{k, v}` |
| `[]any::json_str()` | `string` | JSON 序列化 |

**使用示例：**
```
arr = [5, 3, 1, 4, 2]
arr::sort({a,b} => a < b)          # [1,2,3,4,5]
arr::slice(1, 3)                   # [2,3]
arr::filter({v} => v * 10)         # [10,20,30,40,50]
arr::all({v} => v > 2)             # [3,4,5]
arr::join(',')                     # "1,2,3,4,5"
```

---

#### map[string]any Map 方法

| 方法签名 | 返回类型 | 说明 |
|---|---|---|
| `map::set(key string, val any)` | `map[string]any` | 设置键值，返回自身（可链式） |
| `map::get(key string)` | `any` | 获取键值 |
| `map::len()` | `number` | 键的数量 |
| `map::delete(key string)` | `map[string]any` | 删除键，返回自身 |
| `map::merge(other map)` | `map[string]any` | 合并另一个 map 到自身（就地） |
| `map::clone()` | `map[string]any` | 浅克隆 |
| `map::equals(other map)` | `bool` | 深度相等比较 |
| `map::exclude(keys ...)` | `map[string]any` | 返回排除指定键的新 map |
| `map::some(keys ...)` | `map[string]any` | 返回仅含指定键的新 map |
| `map::keys()` | `[]any` | 返回所有键组成的数组 |
| `map::for(lambda)` | `any` | 遍历，lambda 接收 `{k, v}` |
| `map::to_string(sep1 string, sep2 string)` | `string` | 序列化为 `k=v;k=v` 格式 |
| `map::json_str()` | `string` | JSON 序列化 |

**使用示例：**
```
m = {'a': 1, 'b': 2, 'c': 3}
m::exclude('c')                    # {'a':1,'b':2}
m::some('a', 'b')                  # {'a':1,'b':2}
m::keys()                          # ['a','b','c']
m::to_string('=', ';')             # "a=1;b=2;c=3"
m::set('d', 4)::json_str()
```

---

#### time.Time 时间方法

| 方法签名 | 返回类型 | 说明 |
|---|---|---|
| `time.Time::year()` | `number` | 年份 |
| `time.Time::month()` | `number` | 月份（1-12） |
| `time.Time::day()` | `number` | 日（1-31） |
| `time.Time::hour()` | `number` | 小时（0-23） |
| `time.Time::minute()` | `number` | 分钟（0-59） |
| `time.Time::second()` | `number` | 秒（0-59） |
| `time.Time::unix()` | `number` | 秒级 Unix 时间戳 |
| `time.Time::unix_mill()` | `number` | 毫秒级 Unix 时间戳 |
| `time.Time::unix_micro()` | `number` | 微秒级 Unix 时间戳 |
| `time.Time::format(layout string)` | `string` | 格式化为字符串（Go 时间格式） |
| `time.Time::add_mill(ms number)` | `time.Time` | 增加毫秒数，返回新时间 |
| `time.Time::sub(t time.Time)` | `number` | 两个时间相差的毫秒数 |
| `time.Time::utc()` | `time.Time` | 转换为 UTC 时区 |
| `time.Time::local()` | `time.Time` | 转换为本地时区 |

**使用示例：**
```
now = time_now()
now::year()                        # 2024
now::format('2006-01-02 15:04:05')
now::unix()                        # 秒级时间戳
now::add_mill(3600000)             # 加 1 小时
now::sub(time_parse('2006-01-02', '2024-01-01'))  # 相差毫秒数
```

---

#### float64 数字方法

| 方法签名 | 返回类型 | 说明 |
|---|---|---|
| `number::time()` | `time.Time` | 将秒级 Unix 时间戳转为 `time.Time` |
| `number::json_str()` | `string` | JSON 序列化 |

**使用示例：**
```
t = 1700000000::time()::format('2006-01-02')
```

---

#### []byte 字节切片方法

| 方法签名 | 返回类型 | 说明 |
|---|---|---|
| `[]byte::string()` | `string` | 转为字符串 |
| `[]byte::hex()` | `string` | 转为十六进制字符串 |
| `[]byte::md5()` | `[]byte` | 计算 MD5 |
| `[]byte::base64()` | `string` | Base64 编码 |
| `[]byte::base64d()` | `[]byte` | Base64 解码 |
| `[]byte::slice(a number, b number)` | `[]byte` | 子切片 `[a:b]` |
| `[]byte::copy()` | `[]byte` | 复制字节切片 |

**使用示例：**
```
sig = hmac_sha256('data', 'key')::hex()
raw = base64_decode('aGVsbG8=')::string()   # "hello"
```

---

#### *strings.Builder 字符串构建器方法

| 方法签名 | 返回类型 | 说明 |
|---|---|---|
| `*strings.Builder::write(args ...)` | `*strings.Builder` | 追加字符串（可变参数，链式） |
| `*strings.Builder::string()` | `string` | 获取构建的字符串 |

**使用示例：**
```
result = str_builder()
    ::write('Hello, ')
    ::write(name, '!')
    ::string()
```

---

#### *regexp.Regexp 正则方法

| 方法签名 | 返回类型 | 说明 |
|---|---|---|
| `*regexp.Regexp::match(src string)` | `bool` | 测试字符串是否匹配正则 |

**使用示例：**
```
re  = regexp_new('^[0-9]+$')
ok  = re::match('12345')    # true
ok2 = re::match('abc')      # false
```

---

#### url.Values URL 参数方法

| 方法签名 | 返回类型 | 说明 |
|---|---|---|
| `url.Values::get(key string)` | `string` | 获取参数值 |
| `url.Values::set(key string, val any)` | `url.Values` | 设置参数值，返回自身（可链式） |
| `url.Values::encode()` | `string` | URL 编码为查询字符串 |

**使用示例：**
```
params = url_new_values()
    ::set('page', 1)
    ::set('size', 20)
    ::set('keyword', 'hello world')
qs = params::encode()    # "keyword=hello+world&page=1&size=20"
```

---

#### *atomic.Int64 原子整数方法

| 方法签名 | 返回类型 | 说明 |
|---|---|---|
| `*atomic.Int64::set(n number)` | `*atomic.Int64` | 原子设置值 |
| `*atomic.Int64::add(n number)` | `number` | 原子加法，返回新值 |
| `*atomic.Int64::get()` | `number` | 原子读取值 |

**使用示例：**
```
counter = atomic_int(0)
go({} => counter::add(1))
go({} => counter::add(1))
sleep(10)
total = counter::get()    # 2
```

---

#### *goroutine Goroutine 句柄方法

| 方法签名 | 返回类型 | 说明 |
|---|---|---|
| `*goroutine::join()` | `any` | 阻塞等待 goroutine 完成，返回结果 |
| `*goroutine::resume(data any)` | `*goroutine` | 向 goroutine 发送数据（用于协程间通信） |

**使用示例：**
```
g = go({} => (
    sleep(100);
    'done'
))
result = g::join()    # "done"
```

---

#### *httpResp curl 响应方法

`curl()` 函数返回 `*httpResp`，包含以下字段和方法：

**字段（通过 `.` 访问）：**

| 字段 | 类型 | 说明 |
|---|---|---|
| `.body` | `[]byte` | 响应体（字节切片） |
| `.header` | `map[string]any` | 响应头 |
| `.status` | `number` | HTTP 状态码 |
| `.err` | `any` | 错误信息（nil 表示成功） |

**方法：**

| 方法签名 | 返回类型 | 说明 |
|---|---|---|
| `*httpResp::log(opt map)` | `nil` | 打印响应，opt: `{status:true, header:true, body:true}` |
| `*httpResp::throw()` | `*httpResp` | 非 2xx 或有错误时 panic |
| `*httpResp::failed()` | `string\|nil` | 成功返回 nil，失败返回错误字符串 |

**使用示例：**
```
resp = curl('https://api.example.com/data')
resp::log()                                    # 打印响应体
resp::log({status: true, header: true})        # 打印状态、头、体

# 链式处理
data = curl('https://api.example.com/users')
    ::throw()
    .body
    ::string()
    ::to_json_obj()

# 容错处理
resp = recover({} => curl('https://api.example.com/data'))
if(resp.err == nil).then(
    data = resp.data.body::string()::to_json_obj()
).end()
```

---

#### 通用方法（所有类型均可调用）

以下方法可在任意类型的值上调用（通过 `.` 或 `::` 语法）：

| 方法签名 | 返回类型 | 说明 |
|---|---|---|
| `any::type()` | `string` | 返回值的类型名 |
| `any::string()` | `string` | 转为字符串 |
| `any::boolean()` / `any::bool()` | `bool` | 转为布尔值 |
| `any::number()` | `number` | 转为 float64 |
| `any::to_json_str()` | `string` | JSON 序列化 |
| `any::to_json_obj()` | `any` | JSON 反序列化（对字符串值生效） |
| `any::catch()` | `any` | 若为错误值返回 nil，否则返回自身 |
| `any::unwrap()` | `any` | 若为 `*Result{Err:...}` 则 panic，否则返回 Data |
| `any::recover()` | `*Result` | 执行并捕获 panic |
| `any::is_empty()` | `bool` | 判断是否为 nil/空字符串/空数组 |
| `any::for(lambda)` | `any` | 遍历（对数组/map 有效） |
| `any::go(lambda)` | `*goroutine` | 在新 goroutine 中执行 lambda |
| `any::repeat(n number)` | `[]any` | 重复执行 n 次，收集结果 |
| `any::repeats(n number)` | `nil` | 重复执行 n 次，忽略结果 |
| `any::cost()` | `number` | 测量执行耗时（毫秒） |
| `any::set_to(varName string)` / `any::seto(varName string)` | `any` | 将值赋给变量，同时返回自身 |
| `any::defer(defer_expr lambda)` | `any` | 执行完后调用 defer_expr |
| `any::benchmark()` | `any` | 对值作为 lambda 执行基准测试 |
