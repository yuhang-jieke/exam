package registry

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/consul/api"
)

// Registry Consul服务注册中心
type Registry struct {
	client    *api.Client
	config    *ConsulConfig
	serviceID string
	health    *HealthReporter
	mu        sync.RWMutex
	ready     bool
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewRegistry 创建新的服务注册中心
func NewRegistry(cfg *ConsulConfig) (*Registry, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 创建Consul客户端配置
	consulConfig := api.DefaultConfig()
	consulConfig.Address = cfg.Address
	consulConfig.Scheme = cfg.Scheme
	if cfg.Token != "" {
		consulConfig.Token = cfg.Token
	}

	// 创建Consul客户端
	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	registry := &Registry{
		client: client,
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	// 创建健康报告器
	registry.health = NewHealthReporter(client, cfg)

	return registry, nil
}

// Register 注册服务到Consul
func (r *Registry) Register() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 自动生成服务ID
	if r.config.ServiceID == "" {
		hostname, _ := os.Hostname()
		r.serviceID = fmt.Sprintf("%s-%s-%d", r.config.ServiceName, hostname, r.config.ServicePort)
	} else {
		r.serviceID = r.config.ServiceID
	}

	// 自动检测服务地址
	serviceAddr := r.config.ServiceAddress
	if serviceAddr == "" {
		// 尝试获取本机IP
		addrs, err := net.InterfaceAddrs()
		if err == nil {
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						serviceAddr = ipnet.IP.String()
						break
					}
				}
			}
		}
	}

	// 创建服务注册信息
	registration := &api.AgentServiceRegistration{
		ID:      r.serviceID,
		Name:    r.config.ServiceName,
		Port:    r.config.ServicePort,
		Address: serviceAddr,
		Tags:    r.config.Tags,
		Meta:    r.config.Meta,
		Check: &api.AgentServiceCheck{
			CheckID:                        fmt.Sprintf("%s-health", r.serviceID),
			TTL:                            r.config.TTL,
			DeregisterCriticalServiceAfter: r.config.DeregisterAfter,
		},
	}

	// 注册服务
	if err := r.client.Agent().ServiceRegister(registration); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	log.Printf("[Consul] 服务注册成功: %s (ID: %s) 地址: %s:%d",
		r.config.ServiceName, r.serviceID, serviceAddr, r.config.ServicePort)

	// 启动健康检查
	r.health.SetServiceID(r.serviceID)
	if err := r.health.Start(); err != nil {
		// 健康检查启动失败，注销服务
		_ = r.client.Agent().ServiceDeregister(r.serviceID)
		return fmt.Errorf("failed to start health reporter: %w", err)
	}

	r.ready = true
	return nil
}

// Deregister 从Consul注销服务
func (r *Registry) Deregister() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.ready {
		return nil
	}

	log.Printf("[Consul] 正在注销服务: %s (ID: %s)", r.config.ServiceName, r.serviceID)

	// 停止健康检查
	r.health.Stop()

	// 注销服务
	if err := r.client.Agent().ServiceDeregister(r.serviceID); err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	log.Printf("[Consul] 服务注销成功: %s", r.serviceID)
	r.ready = false
	return nil
}

// GracefulShutdown 优雅退出
func (r *Registry) GracefulShutdown(timeout time.Duration) error {
	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// 等待信号
	sig := <-sigChan
	log.Printf("[Consul] 接收到信号: %v", sig)

	// 设置服务为不健康状态（停止心跳）
	r.health.SetNotReady()
	log.Println("[Consul] 服务已标记为未就绪状态")

	// 等待一段时间让现有请求完成
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 执行注销
	done := make(chan error, 1)
	go func() {
		done <- r.Deregister()
	}()

	select {
	case <-ctx.Done():
		log.Println("[Consul] 优雅退出超时")
		return ctx.Err()
	case err := <-done:
		if err != nil {
			log.Printf("[Consul] 注销服务时出错: %v", err)
		}
		return err
	}
}

// GetClient 获取Consul客户端
func (r *Registry) GetClient() *api.Client {
	return r.client
}

// GetServiceID 获取服务ID
func (r *Registry) GetServiceID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.serviceID
}

// IsReady 检查服务是否就绪
func (r *Registry) IsReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ready
}

// SetNotReady 设置服务为不就绪状态
func (r *Registry) SetNotReady() {
	r.health.SetNotReady()
}

// Close 关闭注册中心
func (r *Registry) Close() error {
	r.cancel()
	return r.Deregister()
}
