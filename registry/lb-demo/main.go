package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/yuhang-jieke/exam/registry"
)

// ServiceInstance 模拟的服务实例
type ServiceInstance struct {
	ID   int
	Port int
	Name string
}

// 启动多个服务实例并注册到Consul
func startMultipleServices(consulAddr string, serviceName string, ports []int) ([]*registry.Client, error) {
	var clients []*registry.Client

	for i, port := range ports {
		// 每个实例使用不同的服务ID
		serviceID := fmt.Sprintf("%s-instance-%d", serviceName, i+1)

		cfg := &registry.ConsulConfig{
			Address:         consulAddr,
			Scheme:          "http",
			ServiceName:     serviceName,
			ServiceID:       serviceID,
			ServicePort:     port,
			TTL:             "10s",
			CheckTimeout:    "5s",
			DeregisterAfter: "30s",
			Tags:            []string{fmt.Sprintf("instance-%d", i+1)},
			Meta: map[string]string{
				"instance_id": fmt.Sprintf("%d", i+1),
				"version":     "1.0.0",
			},
		}

		client, err := registry.NewClient(cfg)
		if err != nil {
			return nil, fmt.Errorf("创建客户端失败 (port %d): %w", port, err)
		}

		if err := client.Register(); err != nil {
			return nil, fmt.Errorf("注册服务失败 (port %d): %w", port, err)
		}

		clients = append(clients, client)
		log.Printf("✅ 服务实例 #%d 注册成功 - 端口: %d, ID: %s", i+1, port, serviceID)
	}

	return clients, nil
}

// 启动简单的HTTP服务器模拟微服务
func startHTTPServer(port int, instanceID int) *http.Server {
	mux := http.NewServeMux()

	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Instance #%d is healthy\n", instanceID)
	})

	// 业务端点
	mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello from Instance #%d (Port: %d)\n", instanceID, port)
	})

	// 信息端点
	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"instance_id": %d, "port": %d, "status": "running"}`, instanceID, port)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		log.Printf("🚀 HTTP服务器启动在端口 %d (实例 #%d)", port, instanceID)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP服务器错误 (端口 %d): %v", port, err)
		}
	}()

	return server
}

// 演示服务发现
func demoServiceDiscovery(client *registry.Client, serviceName string) {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("📡 服务发现演示")
	log.Println(strings.Repeat("=", 60))

	instances, err := client.DiscoverService(serviceName)
	if err != nil {
		log.Printf("❌ 发现服务失败: %v", err)
		return
	}

	log.Printf("🔍 发现 %d 个健康的服务实例:\n", len(instances))
	for i, instance := range instances {
		log.Printf("   [%d] ID: %s, 地址: %s:%d, 标签: %v, 状态: %s",
			i+1, instance.ID, instance.Address, instance.Port, instance.Tags, instance.Health)
	}
}

// 演示负载均衡 - 轮询算法
func demoRoundRobinLB(client *registry.Client, serviceName string, iterations int) {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("⚖️ 负载均衡演示 - 轮询算法 (Round-Robin)")
	log.Println(strings.Repeat("=", 60))

	lb := client.NewLoadBalancer()

	log.Printf("🔄 执行 %d 次服务选择:\n", iterations)
	for i := 1; i <= iterations; i++ {
		instance, err := lb.Select(serviceName)
		if err != nil {
			log.Printf("   [%d] ❌ 选择失败: %v", i, err)
			continue
		}
		log.Printf("   [%d] ✅ 选中: %s:%d (ID: %s)", i, instance.Address, instance.Port, instance.ID)

		// 模拟请求间隔
		time.Sleep(300 * time.Millisecond)
	}
}

// 演示负载均衡 - 随机算法
func demoRandomLB(client *registry.Client, serviceName string, iterations int) {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("⚖️ 负载均衡演示 - 随机算法 (Random)")
	log.Println(strings.Repeat("=", 60))

	lb := client.NewLoadBalancer()

	// 统计每个实例被选中的次数
	stats := make(map[string]int)

	log.Printf("🎲 执行 %d 次随机选择:\n", iterations)
	for i := 1; i <= iterations; i++ {
		instance, err := lb.SelectRandom(serviceName)
		if err != nil {
			log.Printf("   [%d] ❌ 选择失败: %v", i, err)
			continue
		}
		stats[instance.ID]++
	}

	// 打印统计结果
	log.Println("\n📊 选择统计:")
	for id, count := range stats {
		percentage := float64(count) / float64(iterations) * 100
		log.Printf("   %s: %d 次 (%.1f%%)", id, count, percentage)
	}
}

// 演示健康检查 - 模拟服务下线
func demoHealthCheck(consulClient *api.Client, serviceName string) {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🏥 健康检查演示 - 模拟服务下线")
	log.Println(strings.Repeat("=", 60))

	// 获取当前所有服务实例
	services, _, err := consulClient.Health().Service(serviceName, "", false, nil)
	if err != nil {
		log.Printf("❌ 获取服务列表失败: %v", err)
		return
	}

	if len(services) == 0 {
		log.Println("⚠️ 没有发现服务实例")
		return
	}

	// 选择第一个服务实例进行模拟下线
	targetService := services[0].Service
	log.Printf("🔴 模拟下线服务: %s (ID: %s)", targetService.Service, targetService.ID)

	// 设置服务为critical状态
	checkID := fmt.Sprintf("%s-health", targetService.ID)
	err = consulClient.Agent().UpdateTTL(checkID, "Service is going down", "critical")
	if err != nil {
		log.Printf("⚠️ 设置健康检查状态失败: %v (这是正常的，因为TTL检查需要主动心跳)", err)
	}

	log.Println("⏳ 等待健康检查更新...")
	time.Sleep(15 * time.Second)

	// 再次发现服务，看是否排除了不健康的服务
	log.Println("🔍 再次发现服务...")
}

// 演示实时监控服务变化
func demoWatchService(client *registry.Client, serviceName string, duration time.Duration) {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Printf("👁️ 实时监控服务变化 (持续 %v)", duration)
	log.Println(strings.Repeat("=", 60))

	discovery := registry.NewDiscovery(client.GetConsulClient())

	done := make(chan struct{})
	go func() {
		discovery.Watch(serviceName, 3*time.Second, func(updated []*registry.ServiceInstance) {
			log.Printf("\n📌 服务状态更新 (%d 个实例):", len(updated))
			for _, inst := range updated {
				log.Printf("   - %s:%d (状态: %s)", inst.Address, inst.Port, inst.Health)
			}
		})
	}()

	// 等待指定时间
	time.Sleep(duration)
	close(done)
}

// 演示实际HTTP请求负载均衡
func demoHTTPLoadBalancing(client *registry.Client, serviceName string, iterations int) {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🌐 HTTP请求负载均衡演示")
	log.Println(strings.Repeat("=", 60))

	lb := client.NewLoadBalancer()

	log.Printf("📡 发送 %d 次 HTTP 请求:\n", iterations)
	for i := 1; i <= iterations; i++ {
		instance, err := lb.Select(serviceName)
		if err != nil {
			log.Printf("   [%d] ❌ 选择实例失败: %v", i, err)
			continue
		}

		// 构建请求URL
		url := fmt.Sprintf("http://%s:%d/api/hello", instance.Address, instance.Port)

		// 发送HTTP请求
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("   [%d] ❌ HTTP请求失败: %v", i, err)
			continue
		}

		// 读取响应
		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		resp.Body.Close()

		log.Printf("   [%d] ✅ 响应: %s", i, string(buf[:n]))

		time.Sleep(200 * time.Millisecond)
	}
}

// 打印服务统计信息
func printServiceStats(consulClient *api.Client, serviceName string) {
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("📊 Consul服务统计")
	log.Println(strings.Repeat("=", 60))

	// 获取所有健康的服务
	healthyServices, _, err := consulClient.Health().Service(serviceName, "", true, nil)
	if err != nil {
		log.Printf("❌ 获取健康服务失败: %v", err)
		return
	}

	// 获取所有服务（包括不健康的）
	allServices, _, err := consulClient.Health().Service(serviceName, "", false, nil)
	if err != nil {
		log.Printf("❌ 获取所有服务失败: %v", err)
		return
	}

	log.Printf("📈 服务名称: %s", serviceName)
	log.Printf("   健康实例数: %d", len(healthyServices))
	log.Printf("   总实例数: %d", len(allServices))

	if len(allServices) > 0 {
		log.Println("\n📋 实例详情:")
		for _, s := range allServices {
			status := "🟢 健康"
			if s.Checks.AggregatedStatus() != "passing" {
				status = "🔴 不健康"
			}
			log.Printf("   - ID: %s, 地址: %s:%d, 状态: %s",
				s.Service.ID, s.Service.Address, s.Service.Port, status)
		}
	}
}

func main() {
	// 配置
	consulAddr := "localhost:8500"
	serviceName := "demo-service"
	ports := []int{8081, 8082, 8083} // 启动3个服务实例

	log.Println("╔════════════════════════════════════════════════════════════╗")
	log.Println("║     Consul 多服务实例 & 负载均衡演示程序                    ║")
	log.Println("╚════════════════════════════════════════════════════════════╝")

	// 检查Consul是否可用
	log.Printf("\n🔍 检查Consul连接 (%s)...", consulAddr)
	consulConfig := api.DefaultConfig()
	consulConfig.Address = consulAddr
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatalf("❌ 创建Consul客户端失败: %v", err)
	}

	// 尝试连接Consul
	_, err = consulClient.Agent().Self()
	if err != nil {
		log.Fatalf("❌ 无法连接到Consul (%s): %v\n   请确保Consul已启动: consul agent -dev", consulAddr, err)
	}
	log.Println("✅ Consul连接成功")

	// 启动HTTP服务器模拟微服务
	log.Printf("\n🚀 启动 %d 个HTTP服务实例...", len(ports))
	var httpServers []*http.Server
	for i, port := range ports {
		server := startHTTPServer(port, i+1)
		httpServers = append(httpServers, server)
	}
	time.Sleep(500 * time.Millisecond) // 等待服务器启动

	// 注册所有服务实例到Consul
	log.Printf("\n📝 注册服务实例到Consul...")
	clients, err := startMultipleServices(consulAddr, serviceName, ports)
	if err != nil {
		log.Fatalf("❌ 注册服务失败: %v", err)
	}

	// 创建用于服务发现的客户端
	discoveryClient, err := registry.NewClient(&registry.ConsulConfig{
		Address: consulAddr,
	})
	if err != nil {
		log.Fatalf("❌ 创建发现客户端失败: %v", err)
	}

	// 等待服务注册完成
	log.Println("\n⏳ 等待服务注册完成...")
	time.Sleep(3 * time.Second)

	// 打印服务统计
	printServiceStats(consulClient, serviceName)

	// 运行各种演示
	demoServiceDiscovery(discoveryClient, serviceName)
	demoRoundRobinLB(discoveryClient, serviceName, 9)
	demoRandomLB(discoveryClient, serviceName, 30)
	demoHTTPLoadBalancing(discoveryClient, serviceName, 6)

	// 打印最终统计
	printServiceStats(consulClient, serviceName)

	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("🎉 演示完成!")
	log.Println(strings.Repeat("=", 60))
	log.Println("\n💡 提示:")
	log.Println("   - 打开浏览器访问 http://localhost:8500 查看Consul控制台")
	log.Println("   - 你可以看到名为 'demo-service' 的服务有3个实例")
	log.Println("   - 按 Ctrl+C 退出程序并注销所有服务")

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 优雅退出
	log.Println("\n\n🛑 正在关闭服务...")

	// 关闭HTTP服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	for i, server := range httpServers {
		wg.Add(1)
		go func(idx int, s *http.Server) {
			defer wg.Done()
			if err := s.Shutdown(ctx); err != nil {
				log.Printf("⚠️ 关闭HTTP服务器 #%d 失败: %v", idx+1, err)
			} else {
				log.Printf("✅ HTTP服务器 #%d 已关闭", idx+1)
			}
		}(i, server)
	}
	wg.Wait()

	// 注销所有服务
	for i, client := range clients {
		if err := client.Deregister(); err != nil {
			log.Printf("⚠️ 注销服务实例 #%d 失败: %v", i+1, err)
		} else {
			log.Printf("✅ 服务实例 #%d 已注销", i+1)
		}
	}

	log.Println("👋 程序退出")
}
