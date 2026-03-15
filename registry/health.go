package registry

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
)

// HealthReporter TTL健康检查报告器
type HealthReporter struct {
	client    *api.Client
	config    *ConsulConfig
	serviceID string
	checkID   string
	stopChan  chan struct{}
	ready     bool
	mu        sync.Mutex
	running   bool
}

// NewHealthReporter 创建新的健康报告器
func NewHealthReporter(client *api.Client, cfg *ConsulConfig) *HealthReporter {
	return &HealthReporter{
		client:   client,
		config:   cfg,
		stopChan: make(chan struct{}),
		ready:    true,
	}
}

// SetServiceID 设置服务ID
func (h *HealthReporter) SetServiceID(serviceID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.serviceID = serviceID
	h.checkID = fmt.Sprintf("%s-health", serviceID)
}

// Start 启动健康检查报告
func (h *HealthReporter) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.running {
		return nil
	}

	h.running = true
	h.stopChan = make(chan struct{})

	// 启动后台goroutine发送心跳
	go h.heartbeatLoop()

	log.Printf("[Consul] 健康检查已启动，服务: %s (TTL: %s)", h.serviceID, h.config.TTL)
	return nil
}

// Stop 停止健康检查报告
func (h *HealthReporter) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running {
		return
	}

	close(h.stopChan)
	h.running = false

	// 设置服务为不健康状态
	if err := h.reportStatus(false); err != nil {
		log.Printf("[Consul] 设置服务为不健康状态失败: %v", err)
	}

	log.Printf("[Consul] 健康检查已停止，服务: %s", h.serviceID)
}

// SetNotReady 设置服务为不就绪状态（停止心跳）
func (h *HealthReporter) SetNotReady() {
	h.mu.Lock()
	h.ready = false
	h.mu.Unlock()

	// 立即报告服务为不健康状态
	if err := h.reportStatus(false); err != nil {
		log.Printf("[Consul] 报告服务未就绪失败: %v", err)
	}

	log.Printf("[Consul] 服务已标记为未就绪: %s", h.serviceID)
}

// SetReady 设置服务为就绪状态
func (h *HealthReporter) SetReady() {
	h.mu.Lock()
	h.ready = true
	h.mu.Unlock()

	// 立即报告服务为健康状态
	if err := h.reportStatus(true); err != nil {
		log.Printf("[Consul] 报告服务就绪失败: %v", err)
	}

	log.Printf("[Consul] 服务已标记为就绪: %s", h.serviceID)
}

// heartbeatLoop 心跳循环
func (h *HealthReporter) heartbeatLoop() {
	// TTL心跳间隔应该是TTL的一半
	ttlDuration := h.config.GetTTLDuration()
	heartbeatInterval := ttlDuration / 2

	// 如果计算出的间隔小于1秒，则使用1秒
	if heartbeatInterval < time.Second {
		heartbeatInterval = time.Second
	}

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	// 初始立即发送一次心跳
	if err := h.reportStatus(true); err != nil {
		log.Printf("[Consul] 初始心跳失败: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			h.mu.Lock()
			ready := h.ready
			h.mu.Unlock()

			// 只有在ready状态下才发送心跳
			if ready {
				if err := h.reportStatus(true); err != nil {
					log.Printf("[Consul] 心跳失败: %v", err)
					// 心跳失败后重试
					go h.retryHeartbeat()
				}
			}
		case <-h.stopChan:
			return
		}
	}
}

// retryHeartbeat 心跳失败重试
func (h *HealthReporter) retryHeartbeat() {
	maxRetries := 3
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		time.Sleep(retryDelay)

		h.mu.Lock()
		ready := h.ready
		running := h.running
		h.mu.Unlock()

		if !running {
			return
		}

		if ready {
			if err := h.reportStatus(true); err == nil {
				log.Printf("[Consul] 心跳恢复，重试次数: %d", i+1)
				return
			}
		}

		retryDelay *= 2 // 指数退避
	}
}

// reportStatus 报告健康状态
func (h *HealthReporter) reportStatus(healthy bool) error {
	if h.checkID == "" {
		return fmt.Errorf("check ID not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.config.GetCheckTimeoutDuration())
	defer cancel()

	// 使用通道等待API调用完成
	errChan := make(chan error, 1)
	go func() {
		var err error
		if healthy {
			err = h.client.Agent().UpdateTTL(h.checkID, "服务健康", "pass")
		} else {
			err = h.client.Agent().UpdateTTL(h.checkID, "服务不健康", "fail")
		}
		errChan <- err
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("health check timeout")
	case err := <-errChan:
		return err
	}
}

// IsRunning 检查健康报告器是否在运行
func (h *HealthReporter) IsRunning() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.running
}

// IsReady 检查服务是否就绪
func (h *HealthReporter) IsReady() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.ready
}
