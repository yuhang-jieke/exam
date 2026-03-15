# Consul Service Registry

这是一个用于 Consul 服务注册的 Go 库，提供服务注册、健康检查、服务发现和负载均衡功能。

## 特性

- ✅ 服务注册到 Consul
- ✅ TTL 健康检查（每5秒）
- ✅ 服务发现
- ✅ 负载均衡（轮询和随机）
- ✅ 优雅退出
- ✅ 简单易用的 API
- ✅ 支持 go get 安装

## 安装

```bash
go get github.com/yuhang-jieke/opencodeai/registry
```

## 快速开始

### 1. 基本使用

```go
package main

import (
    "log"
    "time"
    "github.com/yuhang-jieke/opencodeai/registry"
)

func main() {
    // 创建配置
    cfg := &registry.ConsulConfig{
        Address:      "localhost:8500",
        ServiceName:  "my-service",
        ServicePort:  8080,
        TTL:          "5s",
        CheckTimeout: "3s",
    }

    // 注册服务
    client, err := registry.SimpleRegister(cfg)
    if err != nil {
        log.Fatalf("Failed to register: %v", err)
    }
    defer client.Close()

    // 等待信号并优雅退出
    registry.WaitSignal(client, 5*time.Second)
}
```

### 2. 快速注册（最简单的方式）

```go
package main

import (
    "log"
    "github.com/yuhang-jieke/opencodeai/registry"
)

func main() {
    // 快速注册服务
    client, err := registry.QuickRegister("my-service", 8080, "localhost:8500")
    if err != nil {
        log.Fatalf("Failed to register: %v", err)
    }
    defer client.Close()

    // 你的服务逻辑...
}
```

### 3. 完整示例（带配置文件）

```go
package main

import (
    "log"
    "time"
    "github.com/spf13/viper"
    "github.com/yuhang-jieke/opencodeai/registry"
)

type Config struct {
    Consul registry.ConsulConfig `mapstructure:"consul"`
}

func main() {
    // 加载配置
    var cfg Config
    viper.SetConfigFile("config.yaml")
    if err := viper.ReadInConfig(); err != nil {
        log.Fatalf("Failed to read config: %v", err)
    }
    if err := viper.Unmarshal(&cfg); err != nil {
        log.Fatalf("Failed to unmarshal config: %v", err)
    }

    // 创建客户端
    client, err := registry.NewClient(&cfg.Consul)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // 注册服务
    if err := client.Register(); err != nil {
        log.Fatalf("Failed to register: %v", err)
    }
    defer client.Close()

    // 你的服务逻辑...

    // 优雅退出
    registry.WaitSignal(client, 5*time.Second)
}
```

### 4. gRPC 服务集成

```go
package main

import (
    "context"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/yuhang-jieke/opencodeai/registry"
    "google.golang.org/grpc"
)

func main() {
    // 注册服务到Consul
    client, err := registry.QuickRegister("grpc-service", 50051, "localhost:8500")
    if err != nil {
        log.Fatalf("Failed to register: %v", err)
    }
    defer client.Close()

    // 启动gRPC服务器
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    // 注册你的gRPC服务...

    // 启动服务器
    go func() {
        if err := grpcServer.Serve(lis); err != nil {
            log.Fatalf("Failed to serve: %v", err)
        }
    }()

    // 等待信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan

    // 优雅退出
    log.Println("Shutting down...")
    
    // 1. 设置服务为不健康
    client.SetNotReady()
    
    // 2. 等待现有请求完成
    time.Sleep(2 * time.Second)
    
    // 3. 停止gRPC服务器
    grpcServer.GracefulStop()
    
    // 4. 从Consul注销
    client.Deregister()
    
    log.Println("Shutdown complete")
}
```

## 服务发现

### 发现服务实例

```go
// 发现所有健康的服务实例
instances, err := client.DiscoverService("target-service")
if err != nil {
    log.Fatalf("Failed to discover service: %v", err)
}

for _, instance := range instances {
    log.Printf("Instance: %s:%d (Health: %s)", 
        instance.Address, instance.Port, instance.Health)
}
```

### 发现单个实例

```go
// 随机选择一个健康的服务实例
instance, err := client.DiscoverOneInstance("target-service")
if err != nil {
    log.Fatalf("Failed to discover instance: %v", err)
}

log.Printf("Selected instance: %s:%d", instance.Address, instance.Port)
```

## 负载均衡

### 轮询负载均衡

```go
// 创建负载均衡器
lb := client.NewLoadBalancer()

// 轮询选择实例
for i := 0; i < 10; i++ {
    instance, err := lb.Select("target-service")
    if err != nil {
        log.Printf("Failed to select instance: %v", err)
        continue
    }
    log.Printf("Selected: %s:%d", instance.Address, instance.Port)
}
```

### 随机负载均衡

```go
// 随机选择实例
instance, err := lb.SelectRandom("target-service")
if err != nil {
    log.Fatalf("Failed to select instance: %v", err)
}
```

## 配置说明

```yaml
Consul:
  address: "localhost:8500"        # Consul服务器地址
  token: ""                        # ACL token（可选）
  scheme: "http"                   # HTTP scheme (http或https)
  service_name: "user-service"     # 服务名称
  service_id: ""                   # 服务ID（为空则自动生成）
  service_port: 8081               # 服务端口
  service_addr: ""                 # 服务地址（为空则自动检测）
  ttl: "5s"                        # TTL持续时间
  check_timeout: "3s"              # 健康检查超时
  deregister_after: "30s"          # 自动注销时间
  tags:                            # 服务标签
    - "grpc"
    - "v1.0"
  meta:                            # 服务元数据
    protocol: "grpc"
    version: "1.0.0"
```

## API 参考

### Client

```go
// 创建客户端
client, err := registry.NewClient(cfg)

// 注册服务
err := client.Register()

// 注销服务
err := client.Deregister()

// 优雅退出
err := client.GracefulShutdown(timeout)

// 设置服务为不健康状态
client.SetNotReady()

// 发现服务
instances, err := client.DiscoverService(name)
instance, err := client.DiscoverOneInstance(name)

// 创建负载均衡器
lb := client.NewLoadBalancer()

// 关闭客户端
err := client.Close()
```

### 快捷函数

```go
// 简化注册
client, err := registry.SimpleRegister(cfg)

// 快速注册
client, err := registry.QuickRegister(serviceName, port, consulAddr)

// 等待信号并优雅退出
registry.WaitSignal(client, timeout)
```

## 健康检查

- **TTL模式**: 服务每5秒发送一次心跳
- **心跳间隔**: TTL的一半（2.5秒）
- **自动重试**: 心跳失败后会自动重试3次
- **优雅退出**: 收到退出信号后立即停止心跳，服务变为不健康状态

## 优雅退出流程

1. 收到信号（SIGINT/SIGTERM/SIGQUIT）
2. 设置服务为不健康状态（停止心跳）
3. 等待一段时间让现有请求完成
4. 从Consul注销服务
5. 关闭服务器

## 注意事项

1. 确保Consul服务器已启动并可访问
2. TTL建议设置为5-10秒，太短会增加网络负担
3. 生产环境建议配置ACL token进行认证
4. 优雅退出时建议等待2-5秒让现有请求完成

## License

MIT