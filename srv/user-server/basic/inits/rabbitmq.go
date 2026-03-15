package inits

import (
	"fmt"
	"log"

	"github.com/yuhang-jieke/exam/mq"
	"github.com/yuhang-jieke/exam/srv/user-server/basic/config"
)

// RabbitMQInit 初始化 RabbitMQ 连接
func RabbitMQInit() error {
	rabbitConf := config.GlobalConf.RabbitMQ

	// 创建生产者
	producer, err := mq.NewProducer(
		rabbitConf.Host,
		rabbitConf.Port,
		rabbitConf.User,
		rabbitConf.Password,
		rabbitConf.VHost,
	)
	if err != nil {
		return fmt.Errorf("failed to create RabbitMQ producer: %w", err)
	}

	// 保存到全局变量
	config.RabbitMQProducer = producer

	// 创建发布器
	config.RabbitMQPublisher = mq.NewPublisher(producer)

	log.Println("[RabbitMQ] 连接成功")
	return nil
}

// RabbitMQClose 关闭 RabbitMQ 连接
func RabbitMQClose() {
	if config.RabbitMQProducer != nil {
		if err := config.RabbitMQProducer.Close(); err != nil {
			log.Printf("[RabbitMQ] 关闭连接失败: %v", err)
		} else {
			log.Println("[RabbitMQ] 连接已关闭")
		}
	}
}

// DeclareQueue 声明队列（便捷方法）
func DeclareQueue(queueName string, durable, autoDelete, exclusive, noWait bool) error {
	if config.RabbitMQProducer == nil {
		return fmt.Errorf("RabbitMQ producer not initialized")
	}
	return config.RabbitMQProducer.DeclareQueue(queueName, durable, autoDelete, exclusive, noWait, nil)
}

// DeclareExchange 声明交换机（便捷方法）
func DeclareExchange(exchangeName, exchangeType string, durable, autoDelete, internal, noWait bool) error {
	if config.RabbitMQProducer == nil {
		return fmt.Errorf("RabbitMQ producer not initialized")
	}
	return config.RabbitMQProducer.DeclareExchange(exchangeName, exchangeType, durable, autoDelete, internal, noWait, nil)
}

// BindQueue 绑定队列到交换机（便捷方法）
func BindQueue(queueName, routingKey, exchangeName string) error {
	if config.RabbitMQProducer == nil {
		return fmt.Errorf("RabbitMQ producer not initialized")
	}
	return config.RabbitMQProducer.BindQueue(queueName, routingKey, exchangeName, false, nil)
}
