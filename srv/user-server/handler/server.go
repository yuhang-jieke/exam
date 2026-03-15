package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	__ "github.com/yuhang-jieke/exam/srv/proto"
	"github.com/yuhang-jieke/exam/srv/user-server/basic/config"
	"github.com/yuhang-jieke/exam/srv/user-server/model"
	"github.com/yuhang-jieke/exam/srv/user-server/pkg"
)

type Server struct {
	__.UnimplementedEcommerceServiceServer
}

// GoodsMessage 商品消息结构
type GoodsMessage struct {
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Stock     int       `json:"stock"`
	Action    string    `json:"action"`    // 操作类型: created, updated, deleted
	Timestamp time.Time `json:"timestamp"` // 消息时间戳
}

// AddGoods 添加商品 - 只发送消息到队列，不写入数据库
func (s *Server) AddGoods(_ context.Context, in *__.AddGoodsReq) (*__.AddGoodsResp, error) {
	// 检查 RabbitMQ 是否初始化
	if config.RabbitMQPublisher == nil {
		log.Println("[RabbitMQ] Publisher not initialized")
		return nil, errors.New("RabbitMQ 未初始化")
	}

	// 构建消息体（不写入数据库，只发送到队列）
	msg := GoodsMessage{
		Name:      in.Name,
		Price:     in.Price,
		Stock:     int(in.Stock),
		Action:    "created",
		Timestamp: time.Now(),
	}

	// 序列化为 JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[RabbitMQ] 序列化商品消息失败: %v", err)
		return nil, err
	}

	// 发送到队列
	err = config.RabbitMQPublisher.PublishJSON(context.Background(), "goods.queue", jsonData)
	if err != nil {
		log.Printf("[RabbitMQ] 发送商品消息失败: %v", err)
		return nil, err
	}

	log.Printf("[RabbitMQ] 商品消息已发送到队列: Name=%s, Price=%.2f, Stock=%d", msg.Name, msg.Price, msg.Stock)

	return &__.AddGoodsResp{
		Message: "商品信息已提交，等待处理",
	}, nil
}

// OrderMessage 订单消息结构
type OrderMessage struct {
	OrderCode string    `json:"order_code"`
	GoodsId   int       `json:"goods_id"`
	Num       int       `json:"num"`
	Price     float64   `json:"price"`
	Action    string    `json:"action"`    // 操作类型：created
	Timestamp time.Time `json:"timestamp"` // 消息时间戳
}

func (s *Server) AddOrder(_ context.Context, in *__.OrderReq) (*__.OrderResp, error) {
	// 检查 RabbitMQ 是否初始化
	if config.RabbitMQPublisher == nil {
		log.Println("[RabbitMQ] Publisher not initialized")
		return nil, errors.New("RabbitMQ 未初始化")
	}

	ordercode := uuid.New().String()[:16]
	var goods model.Goods
	err := goods.GoodsFind(config.DB, in.GoodsId)
	if err != nil {
		return nil, errors.New("商品搜索失败")
	}

	// 检查库存（只预检查，不扣减）
	if in.Num > int64(goods.Stock) {
		return nil, errors.New("库存不足")
	}

	// 构建订单消息（发送到队列，由消费者处理库存扣减）
	msg := OrderMessage{
		OrderCode: ordercode,
		GoodsId:   int(in.GoodsId),
		Num:       int(in.Num),
		Price:     float64(in.Num) * goods.Price,
		Action:    "created",
		Timestamp: time.Now(),
	}

	// 序列化为 JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[RabbitMQ] 序列化订单消息失败：%v", err)
		return nil, err
	}

	// 发送到订单队列
	err = config.RabbitMQPublisher.PublishJSON(context.Background(), "order.queue", jsonData)
	if err != nil {
		log.Printf("[RabbitMQ] 发送订单消息失败：%v", err)
		return nil, err
	}

	log.Printf("[RabbitMQ] 订单消息已发送到队列：OrderCode=%s, GoodsId=%d, Num=%d", msg.OrderCode, msg.GoodsId, msg.Num)
	return &__.OrderResp{
		Message: "订单已提交，等待处理",
	}, nil
}
func (s *Server) DelGoods(_ context.Context, in *__.DelGoodsReq) (*__.DelGoodsResp, error) {
	var goods model.Goods
	err := goods.DeleteGoods(config.DB, in.Id)
	if err != nil {
		return nil, errors.New("删除失败")
	}
	return &__.DelGoodsResp{
		Message: "删除成功",
	}, nil
}
func (s *Server) UpdateGoods(_ context.Context, in *__.UpdateGoodsReq) (*__.UpdateGoodsResp, error) {
	var goods model.Goods
	err := goods.UdateGood(config.DB, in)
	if err != nil {
		return nil, errors.New("修改失败")
	}
	return &__.UpdateGoodsResp{
		Message: "修改成功",
	}, nil
}
func (s *Server) GetGoodsById(_ context.Context, in *__.GetGoodsByIdReq) (*__.GetGoodsByIdResp, error) {
	var goods model.Goods
	err := goods.GoodsFind(config.DB, in.Id)
	if err != nil {
		return nil, errors.New("查找失败")
	}
	return &__.GetGoodsByIdResp{
		Goods: &__.Goods{
			Id:    int64(goods.ID),
			Name:  goods.Name,
			Price: goods.Price,
			Stock: int64(goods.Stock),
		},
	}, nil
}
func (s *Server) AliPay(_ context.Context, in *__.PayReq) (*__.PayResp, error) {
	var order model.Order
	id, err := order.FindId(config.DB, in.Id)
	if err != nil {
		return nil, errors.New("查询失败")
	}
	pay := pkg.AliPay(id.OrderCode, id.Price)
	return &__.PayResp{
		Url: pay,
	}, nil
}
