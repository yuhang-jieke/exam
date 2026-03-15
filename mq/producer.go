package mq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Producer RabbitMQ 生产者封装
type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

// NewProducer 创建新的生产者实例
func NewProducer(host string, port int, user, password, vhost string) (*Producer, error) {
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

	return &Producer{
		conn:    conn,
		channel: ch,
		url:     url,
	}, nil
}

// Close 关闭连接
func (p *Producer) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			log.Printf("[RabbitMQ] 关闭channel失败: %v", err)
		}
	}
	if p.conn != nil {
		if err := p.conn.Close(); err != nil {
			log.Printf("[RabbitMQ] 关闭连接失败: %v", err)
		}
	}
	return nil
}

// DeclareQueue 声明队列
func (p *Producer) DeclareQueue(queueName string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) error {
	_, err := p.channel.QueueDeclare(
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

// DeclareExchange 声明交换机
func (p *Producer) DeclareExchange(exchangeName, exchangeType string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	err := p.channel.ExchangeDeclare(
		exchangeName, // 交换机名称
		exchangeType, // 交换机类型: direct, topic, fanout, headers
		durable,      // 是否持久化
		autoDelete,   // 是否自动删除
		internal,     // 是否内部使用
		noWait,       // 是否等待
		args,         // 额外参数
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}
	return nil
}

// BindQueue 绑定队列到交换机
func (p *Producer) BindQueue(queueName, routingKey, exchangeName string, noWait bool, args amqp.Table) error {
	err := p.channel.QueueBind(
		queueName,    // 队列名称
		routingKey,   // 路由键
		exchangeName, // 交换机名称
		noWait,       // 是否等待
		args,         // 额外参数
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}
	return nil
}

// GetChannel 获取底层channel（用于高级操作）
func (p *Producer) GetChannel() *amqp.Channel {
	return p.channel
}

// GetConnection 获取底层连接
func (p *Producer) GetConnection() *amqp.Connection {
	return p.conn
}
