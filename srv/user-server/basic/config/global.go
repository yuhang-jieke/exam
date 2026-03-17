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
	RDB          *redis.Client
	ConsulClient *registry.Client
	RuntimePort  int

	RabbitMQProducer  *mq.Producer
	RabbitMQPublisher *mq.Publisher

	RuntimeServiceConfig *ServiceConfig
	configMutex          sync.RWMutex

	configCallbacks []func(old, new *ServiceConfig)
	callbacksMutex  sync.RWMutex
)

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

	if oldConfig != nil {
		triggerCallbacks(oldConfig, newConfig)
	}

	log.Printf("[Config] 服务配置已更新: HTTP超时=%ds, gRPC超时=%ds, DB超时=%ds",
		newConfig.HTTPTimeout, newConfig.GRPCTimeout, newConfig.DBTimeout)
}

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
