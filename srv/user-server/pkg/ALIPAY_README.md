# AliPay 使用说明

## 文件结构

```
srv/user-server/pkg/
├── alipay.go          # 核心支付函数（不依赖全局配置）
├── alipay_config.go   # 从全局配置调用的便捷函数
└── alipay_standalone.go # 独立使用示例（已删除）
```

---

## 方式1: 在当前项目内部调用（推荐）

使用全局配置，自动读取 `config.yaml`：

```go
import "github.com/yuhang-jieke/exam/srv/user-server/pkg"

// 使用全局配置
url := pkg.AliPayFromConfig("ORDER123", 99.99)
if url == "" {
    // 处理错误
}
```

---

## 方式2: 在其他项目中调用

传入配置，不依赖全局变量：

```go
import "github.com/yuhang-jieke/exam/srv/user-server/pkg"

// 创建配置
cfg := pkg.AliPayConfig{
    PrivateKey: "your-private-key",
    AppId:      "2021000123456789",
    NotifyURL:  "https://your-domain.com/notify/pay",
    ReturnURL:  "https://your-domain.com/callback",
}

// 调用支付
url := pkg.AliPay(cfg, "ORDER123", 99.99)
```

---

## 方式3: 带错误返回（推荐）

```go
cfg := pkg.AliPayConfig{
    PrivateKey: "your-private-key",
    AppId:      "2021000123456789",
    NotifyURL:  "https://your-domain.com/notify",
    ReturnURL:  "https://your-domain.com/callback",
}

url, err := pkg.AliPayWithResult(cfg, "ORDER123", 99.99)
if err != nil {
    // 处理错误
    log.Printf("支付失败: %v", err)
    return
}
```

---

## 配置说明

| 字段 | 说明 | 必填 |
|-----|------|-----|
| PrivateKey | RSA私钥 | 是 |
| AppId | 支付宝应用ID | 是 |
| NotifyURL | 支付通知回调地址 | 否（建议配置） |
| ReturnURL | 支付完成跳转地址 | 否（建议配置） |

---

## 其他项目使用示例

```go
package main

import (
    "fmt"
    "github.com/yuhang-jieke/exam/srv/user-server/pkg"
)

func main() {
    // 从环境变量或配置中心读取
    cfg := pkg.AliPayConfig{
        PrivateKey: os.Getenv("ALIPAY_PRIVATE_KEY"),
        AppId:      os.Getenv("ALIPAY_APP_ID"),
        NotifyURL:  "https://api.example.com/pay/notify",
        ReturnURL:  "https://www.example.com/pay/callback",
    }

    orderSn := "ORD20260316001"
    amount := 199.99

    url, err := pkg.AliPayWithResult(cfg, orderSn, amount)
    if err != nil {
        fmt.Printf("创建支付失败: %v\n", err)
        return
    }

    fmt.Printf("支付链接: %s\n", url)
}
```