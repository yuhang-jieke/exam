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
	// 创建日志目录
	logDir := "./logs"
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return fmt.Errorf("创建日志目录失败：%w", err)
	}

	// 生成日志文件名（按日期）
	logFileName := filepath.Join(logDir, fmt.Sprintf("consumer_%s.log", time.Now().Format("2006-01-02")))

	// 打开日志文件（追加模式）
	var err error
	logFile, err = os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败：%w", err)
	}

	// 设置日志输出到文件和控制台
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("日志文件初始化成功：%s", logFileName)
	return nil
}

// closeLogger 关闭日志文件
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
var idempotentHandler *mq.IdempotentHandler

func main() {
	// 初始化日志文件
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

	// 2. 创建幂等性处理器
	idempotentHandler = mq.DefaultIdempotentHandler(config.RDB)
	log.Println("[Idempotent] 幂等性处理器初始化成功")

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

	// 4. 创建上下文（用于优雅退出）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 5. 开始消费商品消息
	err = consumer.SubscribeMsg(ctx, "goods.queue", func(body []byte) error {
		return handleGoodsMessage(ctx, body)
	})
	if err != nil {
		log.Fatalf("启动商品消费者失败：%v", err)
	}
	log.Println("[Consumer] 开始监听 goods.queue 队列...")

	// 6. 开始消费订单消息
	err = consumer.SubscribeMsg(ctx, "order.queue", func(body []byte) error {
		return handleOrderMessage(ctx, body)
	})
	if err != nil {
		log.Fatalf("启动订单消费者失败：%v", err)
	}
	log.Println("[Consumer] 开始监听 order.queue 队列...")

	log.Println("[Consumer] 按 Ctrl+C 停止消费")

	// 7. 等待退出信号
	waitForShutdown(cancel)
}

// handleGoodsMessage 处理商品消息（带幂等性检查）
func handleGoodsMessage(ctx context.Context, body []byte) error {
	// 1. 生成消息唯一标识
	messageKey := mq.GenerateMessageKey(body)

	// 2. 幂等性检查
	processed, err := idempotentHandler.IsProcessed(ctx, messageKey)
	if err != nil {
		log.Printf("[Consumer] 幂等性检查失败：%v", err)
		return err
	}

	if processed {
		log.Printf("[Consumer] 消息已处理过，跳过：key=%s", messageKey)
		return nil // 已处理过，直接返回成功
	}

	// 3. 解析消息
	var msg GoodsMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("[Consumer] 解析消息失败：%v", err)
		return err
	}

	log.Printf("[Consumer] 处理商品消息：Name=%s, Price=%.2f, Stock=%d, Key=%s", msg.Name, msg.Price, msg.Stock, messageKey)

	// 4. 根据操作类型处理
	switch msg.Action {
	case "created":
		// 写入数据库（使用全局 config.DB）
		goods := model.Goods{
			Name:  msg.Name,
			Price: msg.Price,
			Stock: msg.Stock,
		}
		if err := config.DB.Create(&goods).Error; err != nil {
			log.Printf("[Consumer] 写入数据库失败：%v", err)
			return err
		}
		log.Printf("[Consumer] 商品已保存到数据库：ID=%d, Name=%s", goods.ID, goods.Name)

	case "updated":
		// 更新逻辑
		log.Printf("[Consumer] 更新商品：Name=%s", msg.Name)

	case "deleted":
		// 删除逻辑
		log.Printf("[Consumer] 删除商品：Name=%s", msg.Name)

	default:
		log.Printf("[Consumer] 未知操作类型：%s", msg.Action)
	}

	return nil
}

// handleOrderMessage 处理订单消息（带幂等性检查）
func handleOrderMessage(ctx context.Context, body []byte) error {
	startTime := time.Now()

	// 1. 生成消息唯一标识
	messageKey := mq.GenerateMessageKey(body)
	log.Printf("[Order] 开始处理订单消息：Key=%s", messageKey)

	// 2. 幂等性检查
	processed, err := idempotentHandler.IsProcessed(ctx, messageKey)
	if err != nil {
		log.Printf("[Order] 幂等性检查失败：%v", err)
		return err
	}

	if processed {
		log.Printf("[Order] 订单已处理过，跳过：key=%s", messageKey)
		return nil
	}
	log.Printf("[Order] 幂等性检查通过：Key=%s", messageKey)

	// 3. 解析消息
	var msg OrderMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("[Order] 解析订单消息失败：%v", err)
		return err
	}
	log.Printf("[Order] 订单详情：OrderCode=%s, GoodsId=%d, Num=%d, Price=%.2f",
		msg.OrderCode, msg.GoodsId, msg.Num, msg.Price)

	// 4. 根据操作类型处理
	switch msg.Action {
	case "created":
		log.Printf("[Order] 开始创建订单记录...")

		// 创建订单
		order := model.Order{
			OrderCode: msg.OrderCode,
			GoodsId:   msg.GoodsId,
			Num:       msg.Num,
			Price:     msg.Price,
		}
		if err := order.OrderAdd(config.DB); err != nil {
			log.Printf("[Order] ❌ 创建订单失败：%v", err)
			return err
		}
		log.Printf("[Order] ✅ 订单创建成功：OrderCode=%s, ID=%d", msg.OrderCode, order.ID)

		// 查询商品信息
		log.Printf("[Order] 开始查询商品信息：GoodsId=%d", msg.GoodsId)
		var goods model.Goods
		if err := goods.GoodsFind(config.DB, int64(msg.GoodsId)); err != nil {
			log.Printf("[Order] ❌ 查询商品失败：%v", err)
			return err
		}
		log.Printf("[Order] 商品查询成功：Name=%s, 当前库存=%d", goods.Name, goods.Stock)

		// 检查库存是否充足
		if goods.Stock < msg.Num {
			log.Printf("[Order] ❌ 库存不足：当前库存=%d, 需要扣减=%d", goods.Stock, msg.Num)
			return fmt.Errorf("库存不足：当前库存=%d, 订单数量=%d", goods.Stock, msg.Num)
		}
		log.Printf("[Order] 库存检查通过：当前库存=%d >= 订单数量=%d", goods.Stock, msg.Num)

		// 扣减商品库存
		log.Printf("[Order] 开始扣减库存：GoodsId=%d, 扣减数量=%d", msg.GoodsId, msg.Num)
		oldStock := goods.Stock

		orderReq := &__.OrderReq{
			GoodsId: int64(msg.GoodsId),
			Num:     int64(msg.Num),
		}
		if err := goods.UdateGoods(config.DB, orderReq); err != nil {
			log.Printf("[Order] ❌ 扣减库存失败：%v", err)
			return err
		}

		newStock := oldStock - msg.Num
		log.Printf("[Order] ✅ 库存扣减成功：GoodsId=%d, 原库存=%d, 扣减=%d, 现库存=%d",
			msg.GoodsId, oldStock, msg.Num, newStock)

		// 记录处理耗时
		elapsed := time.Since(startTime)
		log.Printf("[Order] 🎉 订单处理完成：OrderCode=%s, 总耗时=%v", msg.OrderCode, elapsed)

	default:
		log.Printf("[Order] 未知操作类型：%s", msg.Action)
	}

	return nil
}

// waitForShutdown 等待退出信号
func waitForShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-sigChan
	log.Printf("[Consumer] 收到信号：%v，准备退出...", sig)
	cancel()
	log.Println("[Consumer] 消费者已停止")

	// 关闭日志文件
	closeLogger()
	log.Println("[Consumer] 日志文件已关闭")
}
