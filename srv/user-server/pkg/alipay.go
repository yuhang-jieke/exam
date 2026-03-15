package pkg

import (
	"log"
	"strconv"

	"github.com/smartwalle/alipay/v3"
	"github.com/yuhang-jieke/exam/srv/user-server/basic/config"
)

func AliPay(orderSn string, total float64) string {
	// 检查配置是否加载
	if config.GlobalConf == nil {
		log.Println("[AliPay] 错误：全局配置未初始化")
		return ""
	}

	alipayinit := config.GlobalConf.AliPay

	// 检查必要配置
	if alipayinit.PrivateKey == "" {
		log.Println("[AliPay] 错误：PrivateKey 为空")
		return ""
	}
	if alipayinit.AppId == "" {
		log.Println("[AliPay] 错误：AppId 为空")
		return ""
	}

	var privateKey = alipayinit.PrivateKey // 必须，上一步中使用 RSA 签名验签工具 生成的私钥
	var appId = alipayinit.AppId

	log.Printf("[AliPay] 初始化：AppId=%s", appId)

	client, err := alipay.New(appId, privateKey, false)
	if err != nil {
		log.Printf("[AliPay] 创建客户端失败：%v", err)
		return ""
	}

	var p = alipay.TradeWapPay{}
	p.NotifyURL = alipayinit.NotifyURL
	p.ReturnURL = alipayinit.ReturnURL
	p.Subject = "支付"
	p.OutTradeNo = orderSn
	p.TotalAmount = strconv.FormatFloat(total, 'f', 2, 64)
	p.ProductCode = "QUICK_WAP_WAY"

	log.Printf("[AliPay] 创建支付：订单号=%s, 金额=%.2f", orderSn, total)

	url, err := client.TradeWapPay(p)
	if err != nil {
		log.Printf("[AliPay] 创建支付 URL 失败：%v", err)
		return ""
	}

	// 这个 payURL 即是用于打开支付宝支付页面的 URL，可将输出的内容复制，到浏览器中访问该 URL 即可打开支付页面。
	var payURL = url.String()
	log.Printf("[AliPay] 支付 URL 生成成功：%s", payURL)
	return payURL
}
