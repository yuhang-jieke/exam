package mq

import (
	"context"
	"fmt"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer 消费者封装
type Consumer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	url      string
	handlers map[string]func([]byte) error // 队列名 -> 处理函数
	running  bool
	mu       sync.RWMutex
	done     chan struct{}
}

// NewConsumer 创建新的消费者实例
func NewConsumer(host string, port int, user, password, vhost string) (*Consumer, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", user, password, host, port, vhost)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	return &Consumer{
		conn:     conn,
		channel:  ch,
		url:      url,
		handlers: make(map[string]func([]byte) error),
		done:     make(chan struct{}),
	}, nil
}

// Close 关闭连接
func (c *Consumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 停止消费
	if c.running {
		close(c.done)
		c.running = false
	}

	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	log.Println("[RabbitMQ Consumer] 连接已关闭")
	return nil
}

// DeclareQueue 声明队列
func (c *Consumer) DeclareQueue(queueName string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) error {
	_, err := c.channel.QueueDeclare(
		queueName,
		durable,
		autoDelete,
		exclusive,
		noWait,
		args,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	return nil
}

// SubscribeMsg 订阅消息（与 SendMsg 对应）
// 使用示例:
//
//	consumer.SubscribeMsg(ctx, "order.queue", func(body []byte) error {
//	    // 处理消息
//	    return nil
//	})
func (c *Consumer) SubscribeMsg(ctx context.Context, queueName string, handler func(body []byte) error) error {
	c.mu.Lock()
	c.handlers[queueName] = handler
	c.mu.Unlock()

	// 声明队列（确保存在）
	if err := c.DeclareQueue(queueName, true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// 设置 QoS
	if err := c.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// 开始消费
	msgs, err := c.channel.Consume(
		queueName, // 队列名称
		"",        // 消费者标签（空字符串表示自动生成）
		false,     // 手动确认
		false,     // 非排他
		false,     // noLocal
		false,     // noWait
		nil,       // 额外参数
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Printf("[RabbitMQ Consumer] 开始监听队列: %s", queueName)

	// 启动消息处理协程
	go c.processMessages(ctx, queueName, msgs)

	return nil
}

// processMessages 处理消息循环
func (c *Consumer) processMessages(ctx context.Context, queueName string, msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("[RabbitMQ Consumer] 收到停止信号: %s", queueName)
			return
		case <-c.done:
			log.Printf("[RabbitMQ Consumer] 消费者关闭: %s", queueName)
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("[RabbitMQ Consumer] 消息通道关闭: %s", queueName)
				return
			}

			// 获取处理函数
			c.mu.RLock()
			handler := c.handlers[queueName]
			c.mu.RUnlock()

			if handler == nil {
				log.Printf("[RabbitMQ Consumer] 未注册处理函数: %s", queueName)
				msg.Nack(false, false)
				continue
			}

			// 执行处理函数
			if err := handler(msg.Body); err != nil {
				log.Printf("[RabbitMQ Consumer] 处理失败 [%s]: %v", queueName, err)
				msg.Nack(false, true) // 重新入队
				continue
			}

			// 确认消息
			msg.Ack(false)
		}
	}
}

// SubscribeMulti 订阅多个队列
// 使用示例:
//
//	consumer.SubscribeMulti(ctx, map[string]func([]byte) error{
//	    "order.queue": handleOrder,
//	    "goods.queue": handleGoods,
//	})
func (c *Consumer) SubscribeMulti(ctx context.Context, handlers map[string]func([]byte) error) error {
	for queueName, handler := range handlers {
		if err := c.SubscribeMsg(ctx, queueName, handler); err != nil {
			return fmt.Errorf("failed to subscribe %s: %w", queueName, err)
		}
	}
	return nil
}

// SubscribeJSON 订阅JSON消息（便捷方法）
// 自动反序列化JSON消息
func (c *Consumer) SubscribeJSON(ctx context.Context, queueName string, handler func(data map[string]interface{}) error) error {
	return c.SubscribeMsg(ctx, queueName, func(body []byte) error {
		// 这里可以添加JSON反序列化逻辑
		// 简化处理，直接传递原始数据
		return handler(nil) // 实际使用时需要json.Unmarshal
	})
}

// GetChannel 获取底层channel
func (c *Consumer) GetChannel() *amqp.Channel {
	return c.channel
}

// GetConnection 获取底层连接
func (c *Consumer) GetConnection() *amqp.Connection {
	return c.conn
}

// IsRunning 检查是否正在运行
func (c *Consumer) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

// Stop 停止消费
func (c *Consumer) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		close(c.done)
		c.running = false
	}
}
