package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ================== 便捷函数（推荐使用） ==================
// 注意：这些函数需要传入 publisher 实例，避免循环导入

// SendMessage 发送消息到队列（简单模式）
func SendMessage(publisher *Publisher, queueName string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return publisher.PublishMessage(ctx, queueName, body, "text/plain")
}

// SendJSON 发送 JSON 消息到队列
func SendJSON(publisher *Publisher, queueName string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return publisher.PublishJSON(ctx, queueName, jsonData)
}

// SendToExchange 发送消息到交换机（发布/订阅模式）
func SendToExchange(publisher *Publisher, exchangeName, routingKey string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return publisher.PublishToExchange(ctx, exchangeName, routingKey, jsonData, "application/json")
}

// SendDelayed 发送延迟消息（需要安装延迟插件）
func SendDelayed(publisher *Publisher, exchangeName string, data interface{}, delayMs int64) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return publisher.PublishDelayedMessage(ctx, exchangeName, jsonData, delayMs, "application/json")
}

// ================== 使用示例 ==================

/*
使用方法示例：

1. 在业务代码中导入：
   import (
       "github.com/yuhang-jieke/exam/mq"
       "github.com/yuhang-jieke/exam/srv/user-server/basic/config"
   )

2. 发送消息到队列：
   err := mq.SendJSON(config.RabbitMQPublisher, "order.queue", orderData)

3. 发送消息到交换机：
   err := mq.SendToExchange(config.RabbitMQPublisher, "order.exchange", "order.created", orderData)

4. 发送延迟消息：
   err := mq.SendDelayed(config.RabbitMQPublisher, "delayed.exchange", orderData, 30*60*1000)

5. 初始化队列和交换机：
   err := config.RabbitMQProducer.DeclareQueue("order.queue", true, false, false, false, nil)
   err := config.RabbitMQProducer.DeclareExchange("order.exchange", "topic", true, false, false, false, nil)
   err := config.RabbitMQProducer.BindQueue("order.queue", "order.created", "order.exchange", false, nil)
*/

// ================== 完整业务示例 ==================

// OrderMessage 订单消息结构
type OrderMessage struct {
	OrderID   string  `json:"order_id"`
	UserID    string  `json:"user_id"`
	Total     float64 `json:"total"`
	CreatedAt int64   `json:"created_at"`
}

// PaymentMessage 支付消息结构
type PaymentMessage struct {
	OrderID   string `json:"order_id"`
	PaidAt    int64  `json:"paid_at"`
	PayMethod string `json:"pay_method"`
}

// SendOrderCreated 发送订单创建消息
func SendOrderCreated(publisher *Publisher, orderID, userID string, total float64) error {
	msg := OrderMessage{
		OrderID:   orderID,
		UserID:    userID,
		Total:     total,
		CreatedAt: time.Now().Unix(),
	}
	return SendJSON(publisher, "order.created.queue", msg)
}

// SendOrderPaid 发送订单支付消息
func SendOrderPaid(publisher *Publisher, orderID, payMethod string) error {
	msg := PaymentMessage{
		OrderID:   orderID,
		PaidAt:    time.Now().Unix(),
		PayMethod: payMethod,
	}
	return SendToExchange(publisher, "order.exchange", "order.paid", msg)
}

// SendOrderCancelled 发送订单取消消息
func SendOrderCancelled(publisher *Publisher, orderID, reason string) error {
	msg := map[string]interface{}{
		"order_id":     orderID,
		"cancelled_at": time.Now().Unix(),
		"reason":       reason,
	}
	return SendToExchange(publisher, "order.exchange", "order.cancelled", msg)
}

// SendOrderDelayedCheck 发送订单延迟检查消息
func SendOrderDelayedCheck(publisher *Publisher, orderID string, delayMinutes int) error {
	msg := map[string]interface{}{
		"order_id": orderID,
		"action":   "check_payment",
	}
	delayMs := int64(delayMinutes * 60 * 1000)
	return SendDelayed(publisher, "delayed.exchange", msg, delayMs)
}

// ExampleUsage 使用示例
func ExampleUsage() {
	/*
		// ================== 订单创建消息示例 ==================

		// 订单数据
		order := map[string]interface{}{
			"order_id":   "ORD20260313001",
			"user_id":    "USER123",
			"total":      299.00,
			"created_at": time.Now().Unix(),
		}

		// 发送到订单队列
		err := SendJSON(config.RabbitMQPublisher, "order.created.queue", order)
		if err != nil {
			log.Printf("发送订单消息失败: %v", err)
			return
		}
		log.Println("订单消息发送成功")

		// ================== 订单支付消息示例（使用交换机） ==================

		payment := map[string]interface{}{
			"order_id":   "ORD20260313001",
			"paid_at":    time.Now().Unix(),
			"pay_method": "alipay",
		}

		// 发送到订单交换机，路由键为 order.paid
		err = SendToExchange(config.RabbitMQPublisher, "order.exchange", "order.paid", payment)
		if err != nil {
			log.Printf("发送支付消息失败: %v", err)
			return
		}
		log.Println("支付消息发送成功")

		// ================== 延迟消息示例 ==================

		// 30分钟后检查订单状态
		orderID := "ORD20260313001"
		delayMs := int64(30 * 60 * 1000) // 30分钟

		data := map[string]interface{}{
			"order_id": orderID,
			"action":   "check_payment",
		}

		err = SendDelayed(config.RabbitMQPublisher, "delayed.exchange", data, delayMs)
		if err != nil {
			log.Printf("发送延迟消息失败: %v", err)
			return
		}
		log.Printf("延迟消息发送成功，将在 %d 毫秒后处理", delayMs)

		// ================== 队列和交换机初始化示例 ==================

		// 1. 声明交换机
		err = config.RabbitMQProducer.DeclareExchange(
			"order.exchange", // 交换机名称
			"topic",          // 交换机类型
			true,             // 持久化
			false,            // 不自动删除
			false,            // 不内部使用
			false,            // 不等待
			nil,              // 无额外参数
		)
		if err != nil {
			log.Fatalf("声明交换机失败: %v", err)
		}

		// 2. 声明队列
		err = config.RabbitMQProducer.DeclareQueue(
			"order.created.queue", // 队列名称
			true,                  // 持久化
			false,                 // 不自动删除
			false,                 // 不排他
			false,                 // 不等待
			nil,                   // 无额外参数
		)
		if err != nil {
			log.Fatalf("声明队列失败: %v", err)
		}

		// 3. 绑定队列到交换机
		err = config.RabbitMQProducer.BindQueue(
			"order.created.queue", // 队列名称
			"order.created",       // 路由键
			"order.exchange",      // 交换机名称
			false,                 // 不等待
			nil,                   // 无额外参数
		)
		if err != nil {
			log.Fatalf("绑定队列失败: %v", err)
		}

		log.Println("订单队列初始化成功")
	*/
}
