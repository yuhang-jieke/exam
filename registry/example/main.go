package main

import (
	"log"
	"time"

	"github.com/yuhang-jieke/exam/registry"
)

// 示例1: 最简单的使用方式
func example1() {
	// 快速注册服务
	client, err := registry.QuickRegister("my-service", 8080, "localhost:8500")
	if err != nil {
		log.Fatalf("Failed to register: %v", err)
	}
	defer client.Close()

	// 你的服务逻辑...
	log.Println("Service is running...")

	// 等待信号并优雅退出
	registry.WaitSignal(client, 5*time.Second)
}

// 示例2: 使用配置文件
func example2() {
	cfg := &registry.ConsulConfig{
		Address:         "localhost:8500",
		Scheme:          "http",
		ServiceName:     "my-service",
		ServicePort:     8080,
		TTL:             "5s",
		CheckTimeout:    "3s",
		DeregisterAfter: "30s",
		Tags:            []string{"grpc", "v1.0"},
		Meta:            map[string]string{"version": "1.0.0"},
	}

	client, err := registry.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	if err := client.Register(); err != nil {
		log.Fatalf("Failed to register: %v", err)
	}
	defer client.Close()

	// 你的服务逻辑...
	log.Println("Service is running...")

	// 优雅退出
	registry.WaitSignal(client, 5*time.Second)
}

// 示例3: 服务发现
func example3() {
	// 创建客户端（不需要注册服务）
	cfg := &registry.ConsulConfig{
		Address: "localhost:8500",
	}

	client, err := registry.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 发现服务
	instances, err := client.DiscoverService("target-service")
	if err != nil {
		log.Fatalf("Failed to discover service: %v", err)
	}

	for _, instance := range instances {
		log.Printf("Found instance: %s:%d (Health: %s)",
			instance.Address, instance.Port, instance.Health)
	}
}

// 示例4: 负载均衡
func example4() {
	cfg := &registry.ConsulConfig{
		Address: "localhost:8500",
	}

	client, err := registry.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 创建负载均衡器
	lb := client.NewLoadBalancer()

	// 轮询选择实例
	for i := 0; i < 5; i++ {
		instance, err := lb.Select("target-service")
		if err != nil {
			log.Printf("Failed to select instance: %v", err)
			continue
		}
		log.Printf("Selected instance: %s:%d", instance.Address, instance.Port)
	}
}

// 示例5: 手动控制健康状态
func example5() {
	cfg := &registry.ConsulConfig{
		Address:     "localhost:8500",
		ServiceName: "my-service",
		ServicePort: 8080,
		TTL:         "5s",
	}

	client, err := registry.SimpleRegister(cfg)
	if err != nil {
		log.Fatalf("Failed to register: %v", err)
	}
	defer client.Close()

	// 服务运行中...
	log.Println("Service is running...")

	// 当服务需要维护时，设置为不健康
	// client.SetNotReady()

	// 维护完成后，设置为健康
	// client.SetReady()
}

func main() {
	log.Println("Choose an example to run:")
	log.Println("1. Quick register")
	log.Println("2. Register with config")
	log.Println("3. Service discovery")
	log.Println("4. Load balancing")
	log.Println("5. Manual health control")

	// 运行示例1
	example1()
}
