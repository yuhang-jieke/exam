package mq

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer 消费者封装
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
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
		conn:    conn,
		channel: ch,
		url:     url,
	}, nil
}

// Close 关闭连接
func (c *Consumer) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("[RabbitMQ Consumer] 关闭channel失败: %v", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("[RabbitMQ Consumer] 关闭连接失败: %v", err)
		}
	}
	return nil
}

// DeclareQueue 声明队列
func (c *Consumer) DeclareQueue(queueName string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) error {
	_, err := c.channel.QueueDeclare(
		queueName,  // 队列名称
		durable,    // 是否持久化
		autoDelete, // 是否自动删除
		exclusive,  // 是否排他
		noWait,     // 是否等待
		args,       // 额外参数
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	return nil
}

// Consume 消费消息
// queueName: 队列名称
// consumerTag: 消费者标签
// autoAck: 是否自动确认
// handler: 消息处理函数
func (c *Consumer) Consume(ctx context.Context, queueName, consumerTag string, autoAck bool, handler func(body []byte) error) error {
	// 声明队列（确保队列存在）
	err := c.DeclareQueue(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// 设置 QoS（预取数量）
	err = c.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// 开始消费
	msgs, err := c.channel.Consume(
		queueName,   // 队列名称
		consumerTag, // 消费者标签
		autoAck,     // 是否自动确认
		false,       // 是否排他
		false,       // noLocal
		false,       // noWait
		nil,         // 额外参数
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("[RabbitMQ Consumer] 开始消费队列: %s, consumerTag: %s", queueName, consumerTag)

	// 处理消息
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("[RabbitMQ Consumer] 收到停止信号，停止消费")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("[RabbitMQ Consumer] 消息通道已关闭")
					return
				}

				log.Printf("[RabbitMQ Consumer] 收到消息: %s", string(msg.Body))

				// 调用处理函数
				err := handler(msg.Body)
				if err != nil {
					log.Printf("[RabbitMQ Consumer] 处理消息失败: %v", err)
					// 拒绝消息，重新入队
					if !autoAck {
						msg.Nack(false, true) // requeue = true
					}
					continue
				}

				// 手动确认消息
				if !autoAck {
					msg.Ack(false)
				}

				log.Printf("[RabbitMQ Consumer] 消息处理成功")
			}
		}
	}()

	return nil
}

// SubscribeMsg 订阅消息（便捷方法）
// topic: 队列名称
// handler: 消息处理函数
func (c *Consumer) SubscribeMsg(ctx context.Context, topic string, handler func(body []byte) error) error {
	return c.Consume(ctx, topic, "", false, handler)
}

// GetChannel 获取底层channel
func (c *Consumer) GetChannel() *amqp.Channel {
	return c.channel
}

// GetConnection 获取底层连接
func (c *Consumer) GetConnection() *amqp.Connection {
	return c.conn
}
