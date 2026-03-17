# RabbitMQ 封装模块

简洁的 RabbitMQ 生产者和消费者封装，提供 `SendMsg` 和 `SubscribeMsg` 对应的 API。

## 快速使用

### 生产者 - 发送消息 `SendMsg`

```go
import (
    "context"
    "encoding/json"
    "github.com/yuhang-jieke/exam/srv/user-server/basic/config"
)

// 构建消息
msg := map[string]interface{}{
    "order_code": "ORD123",
    "goods_id":   1,
    "num":        2,
}

// 序列化
jsonData, _ := json.Marshal(msg)

// 发送到队列
err := config.RabbitMQPublisher.SendMsg(ctx, "order.queue", jsonData)
```

### 消费者 - 订阅消息 `SubscribeMsg`

```go
import (
    "context"
    "encoding/json"
    "github.com/yuhang-jieke/exam/mq"
)

// 创建消费者
consumer, _ := mq.NewConsumer("localhost", 5672, "guest", "guest", "/")
defer consumer.Close()

// 订阅消息（与 SendMsg 对应）
consumer.SubscribeMsg(ctx, "order.queue", func(body []byte) error {
    var msg map[string]interface{}
    json.Unmarshal(body, &msg)
    
    // 处理消息...
    return nil
})
```

---

## API 列表

### 生产者 (Publisher)

| 方法 | 说明 |
|-----|------|
| `SendMsg(ctx, queue, body)` | 发送消息到队列（最简单） |
| `PublishMessage(ctx, queue, body, contentType)` | 发送消息（指定类型） |
| `PublishToExchange(ctx, exchange, routingKey, body, contentType)` | 发送到交换机 |

### 消费者 (Consumer)

| 方法 | 说明 |
|-----|------|
| `SubscribeMsg(ctx, queue, handler)` | 订阅队列消息（与 SendMsg 对应） |
| `SubscribeMulti(ctx, handlers)` | 订阅多个队列 |
| `Close()` | 关闭消费者 |

---

## 对应关系

```
生产者                              消费者
─────────────────────────────────────────────────
SendMsg("order.queue", msg)   →   SubscribeMsg("order.queue", handler)
SendMsg("goods.queue", msg)   →   SubscribeMsg("goods.queue", handler)
```