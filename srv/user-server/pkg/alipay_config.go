package pkg

import (
	"log"

	"github.com/yuhang-jieke/exam/srv/user-server/basic/config"
)

// AliPayFromConfig 从全局配置创建支付宝支付链接
// 使用全局 config.GlobalConf，适用于当前项目内部调用
//
// 使用示例:
//
//	// 在 user-server 项目内部
//	url := pkg.AliPayFromConfig("ORDER123", 99.99)
func AliPayFromConfig(orderSn string, total float64) string {
	// 检查全局配置
	if config.GlobalConf == nil {
		log.Println("[AliPay] 错误：全局配置未初始化，请检查 config.yaml 是否存在")
		return ""
	}

	// 转换配置
	cfg := AliPayConfig{
		PrivateKey: config.GlobalConf.AliPay.PrivateKey,
		AppId:      config.GlobalConf.AliPay.AppId,
		NotifyURL:  config.GlobalConf.AliPay.NotifyURL,
		ReturnURL:  config.GlobalConf.AliPay.ReturnURL,
	}

	// 检查配置
	if cfg.PrivateKey == "" {
		log.Println("[AliPay] 错误：请在 config.yaml 中配置 AliPay.PrivateKey")
		return ""
	}
	if cfg.AppId == "" {
		log.Println("[AliPay] 错误：请在 config.yaml 中配置 AliPay.AppId")
		return ""
	}

	return AliPay(cfg, orderSn, total)
}

// AliPayFromConfigWithResult 从全局配置创建支付宝支付链接（带错误返回）
func AliPayFromConfigWithResult(orderSn string, total float64) (string, error) {
	if config.GlobalConf == nil {
		return "", &AliPayError{Code: ErrConfigEmpty, Msg: "全局配置未初始化"}
	}

	cfg := AliPayConfig{
		PrivateKey: config.GlobalConf.AliPay.PrivateKey,
		AppId:      config.GlobalConf.AliPay.AppId,
		NotifyURL:  config.GlobalConf.AliPay.NotifyURL,
		ReturnURL:  config.GlobalConf.AliPay.ReturnURL,
	}

	return AliPayWithResult(cfg, orderSn, total)
}
