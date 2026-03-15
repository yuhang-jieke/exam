package mq

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher 消息发布器，封装消息入队列功能
type Publisher struct {
	producer *Producer
}

// NewPublisher 创建新的发布器
func NewPublisher(producer *Producer) *Publisher {
	return &Publisher{
		producer: producer,
	}
}

// PublishMessage 发布消息到队列（简单模式）
func (p *Publisher) PublishMessage(ctx context.Context, queueName string, body []byte, contentType string) error {
	err := p.producer.channel.PublishWithContext(
		ctx,
		"",        // 交换机名称（空字符串表示默认交换机）
		queueName, // 路由键（队列名称）
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:  contentType,
			Body:         body,
			DeliveryMode: amqp.Persistent, // 消息持久化
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message to queue %s: %w", queueName, err)
	}
	log.Printf("[RabbitMQ] 消息已发送到队列: %s, 大小: %d bytes", queueName, len(body))
	return nil
}

// PublishToExchange 发布消息到交换机
func (p *Publisher) PublishToExchange(ctx context.Context, exchangeName, routingKey string, body []byte, contentType string) error {
	err := p.producer.channel.PublishWithContext(
		ctx,
		exchangeName, // 交换机名称
		routingKey,   // 路由键
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  contentType,
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message to exchange %s: %w", exchangeName, err)
	}
	log.Printf("[RabbitMQ] 消息已发送到交换机: %s, 路由键: %s", exchangeName, routingKey)
	return nil
}

// PublishWithOptions 带完整选项的消息发布
func (p *Publisher) PublishWithOptions(ctx context.Context, exchange, routingKey string, body []byte, opts PublishOptions) error {
	publishing := amqp.Publishing{
		ContentType:  opts.ContentType,
		Body:         body,
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
	}

	// 设置消息ID
	if opts.MessageID != "" {
		publishing.MessageId = opts.MessageID
	}

	// 设置关联ID
	if opts.CorrelationID != "" {
		publishing.CorrelationId = opts.CorrelationID
	}

	// 设置过期时间
	if opts.Expiration != "" {
		publishing.Expiration = opts.Expiration
	}

	// 设置优先级
	if opts.Priority > 0 {
		publishing.Priority = opts.Priority
	}

	// 设置自定义headers
	if opts.Headers != nil {
		publishing.Headers = opts.Headers
	}

	// 设置回复队列
	if opts.ReplyTo != "" {
		publishing.ReplyTo = opts.ReplyTo
	}

	err := p.producer.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		opts.Mandatory,
		opts.Immediate,
		publishing,
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	log.Printf("[RabbitMQ] 消息已发送: exchange=%s, routingKey=%s, messageID=%s", exchange, routingKey, opts.MessageID)
	return nil
}

// PublishOptions 发布选项
type PublishOptions struct {
	ContentType   string     // 内容类型，如 "application/json", "text/plain"
	MessageID     string     // 消息ID
	CorrelationID string     // 关联ID，用于RPC模式
	Expiration    string     // 过期时间，如 "60000" (毫秒)
	Priority      uint8      // 优先级 0-9
	Headers       amqp.Table // 自定义headers
	ReplyTo       string     // 回复队列
	Mandatory     bool       // 必须路由到队列
	Immediate     bool       // 立即投递
}

// PublishJSON 发布JSON消息到队列（便捷方法）
func (p *Publisher) PublishJSON(ctx context.Context, queueName string, jsonData []byte) error {
	return p.PublishMessage(ctx, queueName, jsonData, "application/json")
}

// PublishText 发布文本消息到队列（便捷方法）
func (p *Publisher) PublishText(ctx context.Context, queueName string, text string) error {
	return p.PublishMessage(ctx, queueName, []byte(text), "text/plain")
}

// PublishDelayedMessage 发布延迟消息（需要安装 rabbitmq_delayed_message_exchange 插件）
func (p *Publisher) PublishDelayedMessage(ctx context.Context, exchangeName string, body []byte, delayMs int64, contentType string) error {
	err := p.producer.channel.PublishWithContext(
		ctx,
		exchangeName, // 延迟交换机名称
		"",           // 路由键
		false,
		false,
		amqp.Publishing{
			ContentType:  contentType,
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Headers: amqp.Table{
				"x-delay": delayMs, // 延迟时间（毫秒）
			},
			Timestamp: time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish delayed message: %w", err)
	}
	log.Printf("[RabbitMQ] 延迟消息已发送: exchange=%s, delay=%dms", exchangeName, delayMs)
	return nil
}

// Close 关闭发布器
func (p *Publisher) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}
