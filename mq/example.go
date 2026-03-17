package mq

// ================== 使用示例 ==================

/*
==================== 生产者发送消息 ====================

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

// 发送到队列（最简单的方式）
err := config.RabbitMQPublisher.SendMsg(ctx, "order.queue", jsonData)


==================== 消费者订阅消息 ====================

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

    // 处理消息
    log.Printf("收到订单: %v", msg)

    return nil
})


==================== 订阅多个队列 ====================

consumer.SubscribeMulti(ctx, map[string]func([]byte) error{
    "order.queue": func(body []byte) error {
        // 处理订单消息
        return nil
    },
    "goods.queue": func(body []byte) error {
        // 处理商品消息
        return nil
    },
})


==================== 完整消费者示例 ====================

package main

import (
    "context"
    "encoding/json"
    "log"
    "os/signal"
    "syscall"

    "github.com/yuhang-jieke/exam/mq"
    "gorm.io/gorm"
)

func main() {
    // 1. 创建消费者
    consumer, _ := mq.NewConsumer("localhost", 5672, "guest", "guest", "/")
    defer consumer.Close()

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 2. 订阅订单队列
    consumer.SubscribeMsg(ctx, "order.queue", func(body []byte) error {
        var msg OrderMessage
        json.Unmarshal(body, &msg)

        // 写入订单表
        // 扣减库存
        // ...

        log.Printf("订单处理完成: %s", msg.OrderCode)
        return nil
    })

    // 3. 订阅商品队列
    consumer.SubscribeMsg(ctx, "goods.queue", func(body []byte) error {
        var msg GoodsMessage
        json.Unmarshal(body, &msg)

        // 写入商品表
        // ...

        log.Printf("商品处理完成: %s", msg.Name)
        return nil
    })

    log.Println("消费者已启动，等待消息...")

    // 等待退出
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("消费者关闭")
}
*/
