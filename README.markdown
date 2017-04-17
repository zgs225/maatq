MaatQ
===

基于Redis的简单的消息队列。

### 功能

* [x] 指定N个Gorouting并行处理队列信息
* [x] 指定失败重试次数
* [x] 队列实现分组
* [ ] 实现cron队列
* [ ] `Go`，`PHP`和`Python`的客户端
* [ ] 实现队列任务监控

### 实现

往名为`maatq:default`的Redis列表中写入消息。消息遵循以下协议:

``` json
{
    "id": "xxxx-xxxx-xxxx-xxxx",
    "event": "SendEmail",
    "data": {
        "arg1": 1,
        "arg2": false
    },
    "timestamp": 1257894000,
    "try": 0
}
```

消息的应答格式如下

``` json
{
    "success": false,
    "timestamp": 1257894000,
    "error": "Foo error"
}

{
    "success": true,
    "error": "",
    "timestamp": 1257894000,
    "data": {
        "key1": 123
    }
}
```
