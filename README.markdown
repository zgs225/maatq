MaatQ
===

基于Redis的简单的消息队列。

### 功能

* [x] 指定N个Gorouting并行处理队列信息
* [x] 指定失败重试次数
* [x] 队列实现分组
* [x] 实现优先队列
* [x] 实现周期行任务
* [x] 实现Crontab
* [x] HTTP API
* [x] 持久化
* [ ] `Go`，`PHP`和`Python`的客户端
* [ ] 实现队列任务监控

### HTTP API

* 查询调度器任务列表

```
GET /v1/schedular/list
```

* 发布一条消息

```
POST /v1/messages/dispath
{
    "event": "hello",
    "data": "world"
}
```

* 发布一条延迟的消息

```
POST /v1/messages/delay
{
    "event": "hello",
    "data": "world",
    "delay": "3m"
}
```

* 发布一条周期性的消息

```
POST /v1/messages/period
{
    "event": "hello",
    "data": "world",
    "period": 100
}
```

* 发布一条Crontab消息

```
POST /v1/messages/crontab
{
    "event": "hello",
    "data": "world",
    "crontab": "* */2 * * *"
}
```

* 尝试取消一条消息

```
POST /v1/messages/cancel/xxxxx-xxx-xxxx
```

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
