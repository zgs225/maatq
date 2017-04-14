Mataq
===

基于Redis的简单的消息队列。

### 功能

* 指定N个Gorouting并行处理队列信息
* 指定失败重试次数

### 实现

往名为`mataq:default`的Redis列表中写入消息。消息遵循以下协议:

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

### 启动

``` bash
mataqd -try=3 -parallel=4 -addr='127.0.0.1:6379' -debug=false -password=""
```

#### 参数说明

* `try` 一个任务失败后最大重试次数
* `parallel` 同时执行任务的并发数
* `addr` Redis的地址
* `password` Redis密码
* `debug` 是否开启Debug

### 任务

+ `hello` Hello world 任务
+ `mipush` MAMC 小米推送
+ `mamc_huawei_push` MAMC 华为推送
