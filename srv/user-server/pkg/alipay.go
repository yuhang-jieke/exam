package pkg

import (
	"log"
	"strconv"

	"github.com/smartwalle/alipay/v3"
	"github.com/yuhang-jieke/exam/srv/user-server/basic/config"
)

// AliPayConfig 支付宝配置
/*type AliPayConfig struct {
	PrivateKey string
	AppId      string
	NotifyURL  string
	ReturnURL  string
}*/

// AliPay 创建支付宝支付链接（推荐使用）
// 传入配置，不依赖全局变量，可在任何项目中使用
//
// 使用示例:
//
//	cfg := pkg.AliPayConfig{
//	    PrivateKey: "your-private-key",
//	    AppId:      "your-app-id",
//	    NotifyURL:  "https://your-domain/notify",
//	    ReturnURL:  "https://your-domain/callback",
//	}
//	url := pkg.AliPay(cfg, "ORDER123", 99.99)
func AliPay(cfg config.AliPay, orderSn string, total float64) string {
	// 1. 参数验证
	if cfg.PrivateKey == "" {
		log.Println("[AliPay] 错误：PrivateKey 为空")
		return ""
	}
	if cfg.AppId == "" {
		log.Println("[AliPay] 错误：AppId 为空")
		return ""
	}
	if orderSn == "" {
		log.Println("[AliPay] 错误：订单号为空")
		return ""
	}
	if total <= 0 {
		log.Println("[AliPay] 错误：金额必须大于0")
		return ""
	}

	// 2. 创建支付宝客户端
	client, err := alipay.New(cfg.AppId, cfg.PrivateKey, false)
	if err != nil {
		log.Printf("[AliPay] 创建客户端失败：%v (AppId=%s)", err, cfg.AppId)
		return ""
	}

	// 3. 构建支付请求
	var p = alipay.TradeWapPay{}
	p.NotifyURL = cfg.NotifyURL
	p.ReturnURL = cfg.ReturnURL
	p.Subject = "订单支付"
	p.OutTradeNo = orderSn
	p.TotalAmount = strconv.FormatFloat(total, 'f', 2, 64)
	p.ProductCode = "QUICK_WAP_WAY"

	log.Printf("[AliPay] 创建支付请求：订单号=%s, 金额=%.2f", orderSn, total)

	// 4. 获取支付URL
	url, err := client.TradeWapPay(p)
	if err != nil {
		log.Printf("[AliPay] 创建支付URL失败：%v", err)
		return ""
	}

	payURL := url.String()
	log.Printf("[AliPay] 支付URL生成成功")
	return payURL
}

// AliPayWithResult 创建支付宝支付链接（带错误返回）
// 返回支付URL和错误信息，便于调用方处理错误
func AliPayWithResult(cfg config.AliPay, orderSn string, total float64) (string, error) {
	// 参数验证
	if cfg.PrivateKey == "" {
		return "", &AliPayError{Code: ErrConfigEmpty, Msg: "PrivateKey 为空"}
	}
	if cfg.AppId == "" {
		return "", &AliPayError{Code: ErrConfigEmpty, Msg: "AppId 为空"}
	}
	if orderSn == "" {
		return "", &AliPayError{Code: ErrParamEmpty, Msg: "订单号为空"}
	}
	if total <= 0 {
		return "", &AliPayError{Code: ErrParamEmpty, Msg: "金额必须大于0"}
	}

	// 创建客户端
	client, err := alipay.New(cfg.AppId, cfg.PrivateKey, false)
	if err != nil {
		return "", &AliPayError{Code: ErrClientCreate, Msg: err.Error()}
	}

	// 构建请求
	var p = alipay.TradeWapPay{}
	p.NotifyURL = cfg.NotifyURL
	p.ReturnURL = cfg.ReturnURL
	p.Subject = "订单支付"
	p.OutTradeNo = orderSn
	p.TotalAmount = strconv.FormatFloat(total, 'f', 2, 64)
	p.ProductCode = "QUICK_WAP_WAY"

	// 获取URL
	url, err := client.TradeWapPay(p)
	if err != nil {
		return "", &AliPayError{Code: ErrURLCreate, Msg: err.Error()}
	}

	return url.String(), nil
}

// AliPayError 支付宝错误类型
type AliPayError struct {
	Code int
	Msg  string
}

func (e *AliPayError) Error() string {
	return e.Msg
}

// 错误码
const (
	ErrConfigEmpty  = 1001
	ErrParamEmpty   = 1002
	ErrClientCreate = 1003
	ErrURLCreate    = 1004
)
