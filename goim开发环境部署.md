# 获取源码与依赖
## 安装golang与goland
    略

## 安装jdk、kafka、redis
    goim访问kafka、redis均使用的默认配置
    
### Kafka
获取kafka
```
wget http://apache.mirrors.hoobly.com/kafka/2.6.0/kafka_2.13-2.6.0.tgz
tar -zxvf kafka_2.13-2.6.0.tgz 
cd kafka_2.13-2.6.0/
```
启动zookeeper
```
bin/zookeeper-server-start.sh config/zookeeper.properties
```
启动kafka
```
bin/kafka-server-start.sh config/server.properties
```
### redis
直接下载安装启动，略过不表

如果redis设置了密码，需要在`cmd/logic/logic-example.toml`中修改`redis`节，添加`auth = "******"`

### 获取bilibili/discovery源码
```
git clone https://github.com/bilibili/discovery
cd discovery/cmd/discovery
go build
```
默认配置启动discovery服务
```
./discovery -conf discovery-example.toml
```


## 本地运行goim
### 获取goim源码
```
git clone https://github.com/Terry-Mao/goim
```
确保go module启用
```
go env -w GO111MODULE=on
```
获取依赖
```
go mod download
```
### 构建启动
```
make build
make run
```
关注`target`目录下的`comet.log`、`job.log`、`logic.log`，看有无错误日志

除此之外，其他日志位于`/tmp/`目录下,可以分别看到`logic.`、`comet.`、`job.`前缀的日志文件，如
`/tmp/comet.INFO`、`/tmp/comet.WARNING`、`/tmp/comet.ERROR`等
### 如何停止
```
make stop
```

### 运行客户端
修改`examples/javascript/client.js`，注释掉
`var ws = new WebSocket('ws://sh.tony.wiki:3102/sub');`
打开`//var ws = new WebSocket('ws://127.0.0.1:3102/sub');`
```
cd examples/javascript
go build
go run main.go
```
浏览器打开[http://127.0.0.1:1999/](http://127.0.0.1:1999/)

### 模拟云端下发消息
多播
```
curl -d 'This is a mid message~' http://127.0.0.1:3111/goim/push/mids?operation=1000&mids=123
```
聊天室
```
curl -d 'This is a room message...' http://127.0.0.1:3111/goim/push/room?operation=1000&type=live&room=1000
```
广播
```
curl -d 'This is a broadcast message!' http://127.0.0.1:3111/goim/push/all?operation=1000
```
单播
```
curl -d 'This is a unicast message~' http://127.0.0.1:3111/goim/push/keys?operation=1000&keys=cbe2d1b8-85c9-4bae-8333-3612e8ff751f
```
注意：单播的key是动态生成的（`key = uuid.New().String()`），要去redis查找
```
HKEYS mid_123
```
具体代码见`internal/logic/dao/redis.go`中的`AddMapping`方法

key也可以由客户端自己指定，方法是修改`examples/javascript/client.js`中的`var token`，添加` "key":"abc",`

经验证，key重复的时候，同一时刻只能有一个会话在线，存在互踢的逻辑，位于`internal/comet/bucket.go`的`Put`方法
```
if dch := b.chs[ch.Key]; dch != nil {
    dch.Close()
}
```