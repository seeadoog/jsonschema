阅读expr 目录的代码，这是一个规则引擎，给我整理成一个文档，文档分为两个部分
- 开发者文档： 详细列出给开发者暴露的 api ，调用方式


- 内置函数介绍 介绍规则引擎支持的内置函数
介绍内置函数，分为两个部分：函数介绍要有入参数，出参数介绍。和使用说明，
1 全局函数：
例如 time_now()   : 获取当前时间 返回 time.Time 类型

2 对象函数，介绍规则引擎支持的对象类型包含的子函数。例如： 

map[string]interface{}::len() float64  : 返回map 类型的长度

time.Time::unix()float64  : 返回time类型的时间戳，精确到秒

string::split(num float64)[]any  返回string 切割后的数组。
        

你可以补充下你觉得需要写到文档中的部分，将文档写入expr_help.md
