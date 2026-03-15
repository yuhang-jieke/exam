package registry

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"
)

// Client Consul客户端包装器
type Client struct {
	registry  *Registry
	discovery *Discovery
	config    *ConsulConfig
}

// NewClient 创建新的Consul客户端
func NewClient(cfg *ConsulConfig) (*Client, error) {
	registry, err := NewRegistry(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry: %w", err)
	}

	return &Client{
		registry:  registry,
		discovery: NewDiscovery(registry.GetClient()),
		config:    cfg,
	}, nil
}

// Register 注册服务
func (c *Client) Register() error {
	return c.registry.Register()
}

// Deregister 注销服务
func (c *Client) Deregister() error {
	return c.registry.Deregister()
}

// GracefulShutdown 优雅退出
func (c *Client) GracefulShutdown(timeout time.Duration) error {
	return c.registry.GracefulShutdown(timeout)
}

// SetNotReady 设置服务为不就绪状态
func (c *Client) SetNotReady() {
	c.registry.SetNotReady()
}

// DiscoverService 发现服务实例
func (c *Client) DiscoverService(serviceName string) ([]*ServiceInstance, error) {
	return c.discovery.DiscoverService(serviceName)
}

// DiscoverOneInstance 发现一个服务实例
func (c *Client) DiscoverOneInstance(serviceName string) (*ServiceInstance, error) {
	return c.discovery.DiscoverOneInstance(serviceName)
}

// NewLoadBalancer 创建负载均衡器
func (c *Client) NewLoadBalancer() *LoadBalancer {
	return NewLoadBalancer(c.discovery)
}

// GetConsulClient 获取原始Consul客户端
func (c *Client) GetConsulClient() *api.Client {
	return c.registry.GetClient()
}

// Close 关闭客户端
func (c *Client) Close() error {
	return c.registry.Close()
}

// SimpleRegister 简化的服务注册函数
func SimpleRegister(cfg *ConsulConfig) (*Client, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	if err := client.Register(); err != nil {
		return nil, err
	}

	return client, nil
}

// SimpleRegisterWithShutdown 带优雅退出的简化服务注册
func SimpleRegisterWithShutdown(cfg *ConsulConfig, timeout time.Duration) (*Client, error) {
	client, err := SimpleRegister(cfg)
	if err != nil {
		return nil, err
	}

	// 启动goroutine处理信号
	go func() {
		if err := client.GracefulShutdown(timeout); err != nil {
			log.Printf("[Consul] 优雅退出错误: %v", err)
		}
	}()

	return client, nil
}

// WaitSignal 等待信号并优雅退出
func WaitSignal(client *Client, timeout time.Duration) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-sigChan
	log.Printf("[Consul] 接收到信号: %v，开始优雅退出...", sig)

	// 设置服务为不健康状态
	client.SetNotReady()

	// 等待一段时间让现有请求完成
	time.Sleep(2 * time.Second)

	// 注销服务
	if err := client.Deregister(); err != nil {
		log.Printf("[Consul] 注销服务错误: %v", err)
	}

	log.Println("[Consul] 服务关闭完成")
}

// QuickRegister 快速注册服务（最简化的API）
func QuickRegister(serviceName string, port int, consulAddr string) (*Client, error) {
	cfg := DefaultConfig()
	cfg.ServiceName = serviceName
	cfg.ServicePort = port
	if consulAddr != "" {
		cfg.Address = consulAddr
	}

	return SimpleRegister(cfg)
}

// WithContext 带上下文的注册
func (c *Client) WithContext(ctx context.Context) error {
	// 监听上下文取消
	go func() {
		<-ctx.Done()
		log.Println("[Consul] 上下文已取消，正在注销服务...")
		_ = c.Deregister()
	}()

	return c.Register()
}
