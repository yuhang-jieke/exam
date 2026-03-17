# Redis SETNX 消息幂等性

## 原理

使用 Redis `SETNX`（Set if Not eXists）命令实现消息幂等性：

```
SETNX key value
- 返回 1：设置成功（key不存在）→ 消息未处理，可以处理
- 返回 0：设置失败（key已存在）→ 消息已处理，跳过
```

## 核心方法

```go
// idempotent.go

// 尝试获取处理权（SETNX）
func (h *IdempotentHandler) TryAcquire(ctx, messageKey) (bool, error)

// 执行带幂等性检查的处理（推荐）
func (h *IdempotentHandler) ExecuteWithIdempotent(ctx, messageKey, handler) error

// 处理失败时释放标记
func (h *IdempotentHandler) Release(ctx, messageKey) error
```

---

## 使用示例

### 方式1: 自动幂等处理（推荐）

```go
// 消息处理函数
func handleOrderMessage(body []byte) error {
    ctx := context.Background()
    
    // 自动处理幂等性：
    // 1. SETNX 获取处理权
    // 2. 执行 handler
    // 3. 失败则释放标记，允许重试
    return orderHandler.ExecuteWithIdempotentBody(ctx, body, func() error {
        // 业务逻辑
        return processOrder(body)
    })
}
```

### 方式2: 手动控制

```go
func handleOrderMessage(body []byte) error {
    ctx := context.Background()
    messageKey := mq.GenerateMessageKey(body)
    
    // 1. 尝试获取处理权（SETNX）
    acquired, err := handler.TryAcquire(ctx, messageKey)
    if err != nil {
        return err
    }
    
    if !acquired {
        // 消息已处理，跳过
        return nil
    }
    
    // 2. 执行业务逻辑
    if err := processOrder(body); err != nil {
        // 处理失败，释放标记允许重试
        handler.Release(ctx, messageKey)
        return err
    }
    
    return nil
}
```

---

## 流程图

```
消息到达
    │
    ▼
┌─────────────────────┐
│ Redis SETNX         │
│ key = message_hash  │
└─────────────────────┘
    │
    ├── 返回 1 (设置成功) ──▶ 处理消息 ──▶ 成功 ──▶ 标记自动过期
    │                              │
    │                              └── 失败 ──▶ DEL key（释放标记）
    │
    └── 返回 0 (key已存在) ──▶ 跳过（已处理）
```

---

## 配置

```go
// 预设处理器
mq.DefaultIdempotentHandler(rdb)    // 24小时过期
mq.ShortIdempotentHandler(rdb)      // 1小时过期
mq.LongIdempotentHandler(rdb)       // 7天过期
mq.OrderIdempotentHandler(rdb)      // 订单专用，7天过期
mq.GoodsIdempotentHandler(rdb)      // 商品专用，24小时过期

// 自定义
mq.NewIdempotentHandler(rdb, "prefix:", 48*time.Hour)
```

---

## 消费者中的使用

```go
// srv/consumer/main.go

var orderHandler *mq.IdempotentHandler

func main() {
    // 初始化 Redis
    inits.RedisInit()
    
    // 创建幂等性处理器
    orderHandler = mq.OrderIdempotentHandler(config.RDB)
    
    // 订阅消息
    consumer.SubscribeMsg(ctx, "order.queue", handleOrderMessage)
}

func handleOrderMessage(body []byte) error {
    ctx := context.Background()
    
    // Redis SETNX 自动幂等处理
    return orderHandler.ExecuteWithIdempotentBody(ctx, body, func() error {
        // 业务逻辑...
        return nil
    })
}
```

---

## Redis Key 格式

```
订单消息: order:processed:{md5_hash}
商品消息: goods:processed:{md5_hash}
默认:    mq:processed:{md5_hash}
```