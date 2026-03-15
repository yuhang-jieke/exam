package config

import (
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/yuhang-jieke/exam/mq"
	"github.com/yuhang-jieke/exam/registry"
	"gorm.io/gorm"
)

var (
	GlobalConf   *AppConfig
	DB           *gorm.DB
	RDB          *redis.Client // Redis 客户端
	ConsulClient *registry.Client
	RuntimePort  int // 运行时端口（由命令行参数指定）

	// RabbitMQ 生产者和发布器
	RabbitMQProducer  *mq.Producer
	RabbitMQPublisher *mq.Publisher

	// RuntimeServiceConfig 运行时服务配置（支持热更新）
	RuntimeServiceConfig *ServiceConfig
	configMutex          sync.RWMutex

	// 配置变更回调函数列表
	configCallbacks []func(old, new *ServiceConfig)
	callbacksMutex  sync.RWMutex
)

// RegisterConfigCallback 注册配置变更回调函数
func RegisterConfigCallback(callback func(old, new *ServiceConfig)) {
	callbacksMutex.Lock()
	defer callbacksMutex.Unlock()
	configCallbacks = append(configCallbacks, callback)
}

// UpdateServiceConfig 更新服务配置（热更新）
func UpdateServiceConfig(newConfig *ServiceConfig) {
	if newConfig == nil {
		return
	}

	configMutex.Lock()
	var oldConfig *ServiceConfig
	if RuntimeServiceConfig != nil {
		oldCopy := *RuntimeServiceConfig
		oldConfig = &oldCopy
	}
	RuntimeServiceConfig = newConfig
	configMutex.Unlock()

	// 触发配置变更回调
	if oldConfig != nil {
		triggerCallbacks(oldConfig, newConfig)
	}

	log.Printf("[Config] 服务配置已更新: HTTP超时=%ds, gRPC超时=%ds, DB超时=%ds",
		newConfig.HTTPTimeout, newConfig.GRPCTimeout, newConfig.DBTimeout)
}

// triggerCallbacks 触发配置变更回调
func triggerCallbacks(old, new *ServiceConfig) {
	callbacksMutex.RLock()
	callbacks := make([]func(old, new *ServiceConfig), len(configCallbacks))
	copy(callbacks, configCallbacks)
	callbacksMutex.RUnlock()

	for _, callback := range callbacks {
		go func(cb func(old, new *ServiceConfig)) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[Config] 回调函数panic: %v", r)
				}
			}()
			cb(old, new)
		}(callback)
	}
}

// GetServiceConfig 获取当前服务配置
func GetServiceConfig() *ServiceConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	if RuntimeServiceConfig == nil {
		return &ServiceConfig{
			HTTPTimeout:   30,
			GRPCTimeout:   30,
			DBTimeout:     10,
			RedisTimeout:  5,
			MaxRetryCount: 3,
			DebugMode:     false,
		}
	}
	return RuntimeServiceConfig
}

// GetHTTPTimeout 获取HTTP超时时间
func GetHTTPTimeout() int {
	return GetServiceConfig().HTTPTimeout
}

// GetGRPCTimeout 获取gRPC超时时间
func GetGRPCTimeout() int {
	return GetServiceConfig().GRPCTimeout
}

// GetDBTimeout 获取数据库超时时间
func GetDBTimeout() int {
	return GetServiceConfig().DBTimeout
}
