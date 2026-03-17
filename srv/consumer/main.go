package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/yuhang-jieke/exam/mq"
	__ "github.com/yuhang-jieke/exam/srv/proto"
	"github.com/yuhang-jieke/exam/srv/user-server/basic/config"
	"github.com/yuhang-jieke/exam/srv/user-server/basic/inits"
	"github.com/yuhang-jieke/exam/srv/user-server/model"
)

// 全局日志文件句柄
var logFile *os.File

// initLogger 初始化日志文件
func initLogger() error {
	logDir := "./logs"
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return fmt.Errorf("创建日志目录失败：%w", err)
	}

	logFileName := filepath.Join(logDir, fmt.Sprintf("consumer_%s.log", time.Now().Format("2006-01-02")))

	var err error
	logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败：%w", err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("日志文件初始化成功：%s", logFileName)
	return nil
}

func closeLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// GoodsMessage 商品消息结构
type GoodsMessage struct {
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

// OrderMessage 订单消息结构
type OrderMessage struct {
	OrderCode string    `json:"order_code"`
	GoodsId   int       `json:"goods_id"`
	Num       int       `json:"num"`
	Price     float64   `json:"price"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

// 全局幂等性处理器
var orderHandler *mq.IdempotentHandler
var goodsHandler *mq.IdempotentHandler

func main() {
	if err := initLogger(); err != nil {
		log.Fatalf("初始化日志失败：%v", err)
	}
	defer closeLogger()

	log.Println("========== 消费者服务启动 ==========")

	// 1. 初始化配置、数据库、Redis
	inits.ConfigInit()
	inits.MysqlInit()
	inits.RedisInit()
	log.Println("[DB] 数据库连接成功")
	log.Println("[Redis] Redis 连接成功")

	// 2. 创建幂等性处理器（基于 Redis SETNX）
	orderHandler = mq.OrderIdempotentHandler(config.RDB)
	goodsHandler = mq.GoodsIdempotentHandler(config.RDB)
	log.Println("[Idempotent] 幂等性处理器初始化成功（Redis SETNX）")

	// 3. 初始化 RabbitMQ 消费者
	rabbitConf := config.GlobalConf.RabbitMQ
	consumer, err := mq.NewConsumer(
		rabbitConf.Host,
		rabbitConf.Port,
		rabbitConf.User,
		rabbitConf.Password,
		rabbitConf.VHost,
	)
	if err != nil {
		log.Fatalf("RabbitMQ 消费者初始化失败：%v", err)
	}
	defer consumer.Close()
	log.Println("[RabbitMQ] 消费者连接成功")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4. 开始消费商品消息
	err = consumer.SubscribeMsg(ctx, "goods.queue", handleGoodsMessage)
	if err != nil {
		log.Fatalf("启动商品消费者失败：%v", err)
	}
	log.Println("[Consumer] 开始监听 goods.queue 队列...")

	// 5. 开始消费订单消息
	err = consumer.SubscribeMsg(ctx, "order.queue", handleOrderMessage)
	if err != nil {
		log.Fatalf("启动订单消费者失败：%v", err)
	}
	log.Println("[Consumer] 开始监听 order.queue 队列...")

	log.Println("[Consumer] 按 Ctrl+C 停止消费")

	waitForShutdown(cancel)
}

// handleGoodsMessage 处理商品消息（Redis SETNX 幂等性）
func handleGoodsMessage(body []byte) error {
	ctx := context.Background()
	messageKey := mq.GenerateMessageKey(body)

	// 使用 ExecuteWithIdempotent 自动处理幂等性
	// 基于 Redis SETNX 实现：获取处理权 -> 执行 -> 失败则释放
	return goodsHandler.ExecuteWithIdempotentBody(ctx, body, func() error {
		return processGoodsMessage(body, messageKey)
	})
}

// processGoodsMessage 实际的商品消息处理逻辑
func processGoodsMessage(body []byte, messageKey string) error {
	var msg GoodsMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("[Goods] 解析消息失败：%v", err)
		return err
	}

	log.Printf("[Goods] 处理商品消息：Name=%s, Price=%.2f, Stock=%d, Key=%s",
		msg.Name, msg.Price, msg.Stock, messageKey)

	switch msg.Action {
	case "created":
		goods := model.Goods{
			Name:  msg.Name,
			Price: msg.Price,
			Stock: msg.Stock,
		}
		if err := config.DB.Create(&goods).Error; err != nil {
			log.Printf("[Goods] 写入数据库失败：%v", err)
			return err
		}
		log.Printf("[Goods] 商品已保存：ID=%d, Name=%s", goods.ID, goods.Name)

	case "updated":
		log.Printf("[Goods] 更新商品：Name=%s", msg.Name)

	case "deleted":
		log.Printf("[Goods] 删除商品：Name=%s", msg.Name)

	default:
		log.Printf("[Goods] 未知操作类型：%s", msg.Action)
	}

	return nil
}

// handleOrderMessage 处理订单消息（Redis SETNX 幂等性）
func handleOrderMessage(body []byte) error {
	ctx := context.Background()
	messageKey := mq.GenerateMessageKey(body)

	// 使用 ExecuteWithIdempotent 自动处理幂等性
	// Redis SETNX 确保消息只被处理一次
	return orderHandler.ExecuteWithIdempotentBody(ctx, body, func() error {
		return processOrderMessage(body, messageKey)
	})
}

// processOrderMessage 实际的订单消息处理逻辑
func processOrderMessage(body []byte, messageKey string) error {
	startTime := time.Now()
	log.Printf("[Order] 开始处理：Key=%s", messageKey)

	var msg OrderMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("[Order] 解析消息失败：%v", err)
		return err
	}

	log.Printf("[Order] 订单详情：OrderCode=%s, GoodsId=%d, Num=%d, Price=%.2f",
		msg.OrderCode, msg.GoodsId, msg.Num, msg.Price)

	switch msg.Action {
	case "created":
		// 创建订单
		order := model.Order{
			OrderCode: msg.OrderCode,
			GoodsId:   msg.GoodsId,
			Num:       msg.Num,
			Price:     msg.Price,
		}
		if err := order.OrderAdd(config.DB); err != nil {
			log.Printf("[Order] 创建订单失败：%v", err)
			return err
		}
		log.Printf("[Order] 订单创建成功：OrderCode=%s, ID=%d", msg.OrderCode, order.ID)

		// 查询商品
		var goods model.Goods
		if err := goods.GoodsFind(config.DB, int64(msg.GoodsId)); err != nil {
			log.Printf("[Order] 查询商品失败：%v", err)
			return err
		}
		log.Printf("[Order] 商品查询成功：Name=%s, 库存=%d", goods.Name, goods.Stock)

		// 检查库存
		if goods.Stock < msg.Num {
			log.Printf("[Order] 库存不足：当前=%d, 需要=%d", goods.Stock, msg.Num)
			return fmt.Errorf("库存不足")
		}

		// 扣减库存
		orderReq := &__.OrderReq{
			GoodsId: int64(msg.GoodsId),
			Num:     int64(msg.Num),
		}
		if err := goods.UdateGoods(config.DB, orderReq); err != nil {
			log.Printf("[Order] 扣减库存失败：%v", err)
			return err
		}

		elapsed := time.Since(startTime)
		log.Printf("[Order] 处理完成：OrderCode=%s, 耗时=%v", msg.OrderCode, elapsed)

	default:
		log.Printf("[Order] 未知操作类型：%s", msg.Action)
	}

	return nil
}

func waitForShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-sigChan
	log.Printf("[Consumer] 收到信号：%v，准备退出...", sig)
	cancel()
	closeLogger()
	log.Println("[Consumer] 消费者已停止")
}
