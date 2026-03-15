package registry

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

// ServiceInstance 服务实例信息
type ServiceInstance struct {
	ID      string
	Name    string
	Address string
	Port    int
	Tags    []string
	Meta    map[string]string
	Health  string // "passing", "warning", "critical"
}

// Discovery 服务发现
type Discovery struct {
	client *api.Client
	mu     sync.RWMutex
}

// NewDiscovery 创建新的服务发现
func NewDiscovery(client *api.Client) *Discovery {
	return &Discovery{
		client: client,
	}
}

// DiscoverService 发现服务实例
func (d *Discovery) DiscoverService(serviceName string) ([]*ServiceInstance, error) {
	// 查询健康的服务实例
	services, _, err := d.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover service %s: %w", serviceName, err)
	}

	instances := make([]*ServiceInstance, 0, len(services))
	for _, service := range services {
		instance := &ServiceInstance{
			ID:      service.Service.ID,
			Name:    service.Service.Service,
			Address: service.Service.Address,
			Port:    service.Service.Port,
			Tags:    service.Service.Tags,
			Meta:    service.Service.Meta,
		}

		// 获取健康状态
		if len(service.Checks) > 0 {
			instance.Health = service.Checks.AggregatedStatus()
		} else {
			instance.Health = "unknown"
		}

		instances = append(instances, instance)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no healthy instances found for service %s", serviceName)
	}

	return instances, nil
}

// DiscoverOneInstance 发现服务的一个实例（随机选择）
func (d *Discovery) DiscoverOneInstance(serviceName string) (*ServiceInstance, error) {
	instances, err := d.DiscoverService(serviceName)
	if err != nil {
		return nil, err
	}

	// 随机选择一个实例
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return instances[r.Intn(len(instances))], nil
}

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	discovery *Discovery
	instances []*ServiceInstance
	index     int
	mu        sync.Mutex
}

// NewLoadBalancer 创建新的负载均衡器
func NewLoadBalancer(discovery *Discovery) *LoadBalancer {
	return &LoadBalancer{
		discovery: discovery,
	}
}

// Select 使用轮询算法选择一个服务实例
func (lb *LoadBalancer) Select(serviceName string) (*ServiceInstance, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// 如果没有实例或需要刷新，则重新获取
	if len(lb.instances) == 0 {
		instances, err := lb.discovery.DiscoverService(serviceName)
		if err != nil {
			return nil, err
		}
		lb.instances = instances
		lb.index = 0
	}

	// 轮询选择
	instance := lb.instances[lb.index]
	lb.index = (lb.index + 1) % len(lb.instances)

	return instance, nil
}

// SelectRandom 使用随机算法选择一个服务实例
func (lb *LoadBalancer) SelectRandom(serviceName string) (*ServiceInstance, error) {
	return lb.discovery.DiscoverOneInstance(serviceName)
}

// Refresh 刷新服务实例列表
func (lb *LoadBalancer) Refresh(serviceName string) error {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	instances, err := lb.discovery.DiscoverService(serviceName)
	if err != nil {
		return err
	}

	lb.instances = instances
	lb.index = 0
	return nil
}

// Watch 监听服务变化（简单的轮询方式）
func (d *Discovery) Watch(serviceName string, interval time.Duration, callback func(instances []*ServiceInstance)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		instances, err := d.DiscoverService(serviceName)
		if err != nil {
			log.Printf("[Consul] Failed to discover service %s: %v", serviceName, err)
			continue
		}

		callback(instances)
	}
}

// GetServiceAddress 获取服务的完整地址
func (s *ServiceInstance) GetServiceAddress() string {
	return fmt.Sprintf("%s:%d", s.Address, s.Port)
}

// IsHealthy 检查服务实例是否健康
func (s *ServiceInstance) IsHealthy() bool {
	return s.Health == "passing"
}
