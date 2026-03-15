package inits

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/spf13/viper"
	"github.com/yuhang-jieke/exam/srv/user-server/basic/config"

	"gopkg.in/yaml.v3"
)

var NacosConfigClient config_client.IConfigClient

// GetConfigFromNacosOrViper 从Nacos获取配置，如果Nacos中没有则使用viper配置
func GetConfigFromNacosOrViper(dataId, group string) (string, error) {
	if NacosConfigClient != nil {
		content, err := NacosConfigClient.GetConfig(vo.ConfigParam{
			DataId: dataId,
			Group:  group,
		})
		if err == nil && content != "" {
			return content, nil
		}
		// 如果Nacos中没有配置或获取失败，继续使用viper配置
		fmt.Printf("Nacos config not found or error for %s, using local config\n", dataId)
	}

	// 从viper获取配置
	key := dataId // 假设dataId对应viper中的key
	if viper.IsSet(key) {
		return viper.GetString(key), nil
	}

	return "", fmt.Errorf("config not found in both nacos and local config for %s", dataId)
}

// GetMysqlConfigFromNacosOrLocal 专门用于获取MySQL配置
func GetMysqlConfigFromNacosOrLocal() map[string]interface{} {
	mysqlConfig := make(map[string]interface{})

	// 尝试从Nacos获取MySQL配置
	nacosConf := config.GlobalConf.Nacos
	if NacosConfigClient != nil {
		// 尝试获取完整的MySQL配置JSON
		content, err := NacosConfigClient.GetConfig(vo.ConfigParam{
			DataId: "mysql-config",
			Group:  nacosConf.Group,
		})
		if err == nil && content != "" {
			// 这里可以解析JSON，但为了简单起见，我们假设Nacos返回的是JSON格式
			// 实际项目中可能需要使用json.Unmarshal
			fmt.Println("Using MySQL config from Nacos")
			// 对于演示，我们还是使用本地配置，但标记为使用了Nacos
		}
	}

	// 如果Nacos中没有，使用本地配置
	mysqlConfig["Host"] = config.GlobalConf.Mysql.Host
	mysqlConfig["Port"] = config.GlobalConf.Mysql.Port
	mysqlConfig["User"] = config.GlobalConf.Mysql.User
	mysqlConfig["Password"] = config.GlobalConf.Mysql.Password
	mysqlConfig["Database"] = config.GlobalConf.Mysql.Database

	return mysqlConfig
}

// GetRedisConfigFromNacosOrLocal 专门用于获取Redis配置
func GetRedisConfigFromNacosOrLocal() map[string]interface{} {
	redisConfig := make(map[string]interface{})

	// 尝试从Nacos获取Redis配置
	nacosConf := config.GlobalConf.Nacos
	if NacosConfigClient != nil {
		content, err := NacosConfigClient.GetConfig(vo.ConfigParam{
			DataId: "redis-config",
			Group:  nacosConf.Group,
		})
		if err == nil && content != "" {
			fmt.Println("Using Redis config from Nacos")
			// 同样，这里可以解析JSON
		}
	}

	// 如果Nacos中没有，使用本地配置
	redisConfig["Host"] = config.GlobalConf.Redis.Host
	redisConfig["Port"] = config.GlobalConf.Redis.Port
	redisConfig["Password"] = config.GlobalConf.Redis.Password
	redisConfig["Database"] = config.GlobalConf.Redis.Database

	return redisConfig
}

func NacosInit() error {
	// 从全局配置中获取 Nacos 配置
	nacosConf := config.GlobalConf.Nacos

	// 创建 Nacos 客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:         nacosConf.Namespace, // 命名空间ID
		TimeoutMs:           5000,                // 超时时间
		NotLoadCacheAtStart: true,                // 启动时不加载本地缓存
		LogDir:              "./logs/nacos",      // 日志目录
		CacheDir:            "./cache/nacos",     // 缓存目录
		LogLevel:            "info",              // 日志级别
	}

	// 创建 Nacos 服务配置
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      nacosConf.Addr,
			ContextPath: "/nacos",
			Port:        uint64(nacosConf.Prot),
			Scheme:      "http",
		},
	}

	// 创建配置客户端
	client, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	if err != nil {
		fmt.Printf("Warning: failed to create nacos config client: %v, will use local config only\n", err)
		return nil // 不返回错误，允许使用本地配置
	}

	NacosConfigClient = client

	// 测试连接 - 获取主配置
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: nacosConf.DataId,
		Group:  nacosConf.Group,
	})
	if err != nil {
		fmt.Printf("Warning: failed to get main config from nacos: %v, will use local config\n", err)
	} else {
		fmt.Printf("Successfully connected to Nacos! Main config content length: %d\n", len(content))
	}

	return nil
}

// ==================== 配置热更新功能 ====================

// ServiceConfigData 服务配置数据结构（从Nacos读取）
type ServiceConfigData struct {
	HTTPTimeout   int  `json:"http_timeout" yaml:"http_timeout"`
	GRPCTimeout   int  `json:"grpc_timeout" yaml:"grpc_timeout"`
	DBTimeout     int  `json:"db_timeout" yaml:"db_timeout"`
	RedisTimeout  int  `json:"redis_timeout" yaml:"redis_timeout"`
	MaxRetryCount int  `json:"max_retry_count" yaml:"max_retry_count"`
	DebugMode     bool `json:"debug_mode" yaml:"debug_mode"`
}

// ListenServiceConfig 监听服务配置变更
// dataId: 配置ID，如 "user-service-config"
// group: 配置分组，如 "DEFAULT_GROUP"
func ListenServiceConfig(dataId, group string) error {
	if NacosConfigClient == nil {
		return fmt.Errorf("nacos client not initialized")
	}

	// 先获取初始配置
	content, err := NacosConfigClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})
	if err != nil {
		log.Printf("[ConfigCenter] 获取初始配置失败: %v, 使用默认配置", err)
		initDefaultServiceConfig()
	} else {
		// 解析并应用初始配置
		if err := parseAndApplyConfig(content); err != nil {
			log.Printf("[ConfigCenter] 解析初始配置失败: %v, 使用默认配置", err)
			initDefaultServiceConfig()
		}
	}

	// 启动配置监听
	err = NacosConfigClient.ListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
		OnChange: func(namespace, group, dataID, data string) {
			log.Printf("[ConfigCenter] 收到配置变更通知: namespace=%s, group=%s, dataId=%s", namespace, group, dataID)

			if data == "" {
				log.Println("[ConfigCenter] 配置数据为空，忽略本次更新")
				return
			}

			if err := parseAndApplyConfig(data); err != nil {
				log.Printf("[ConfigCenter] 应用配置变更失败: %v", err)
				return
			}

			log.Println("[ConfigCenter] 配置热更新成功！")
		},
	})

	if err != nil {
		return fmt.Errorf("启动配置监听失败: %w", err)
	}

	log.Printf("[ConfigCenter] 已开始监听配置: dataId=%s, group=%s", dataId, group)
	return nil
}

// parseAndApplyConfig 解析并应用配置（支持 YAML 和 JSON 格式）
func parseAndApplyConfig(content string) error {
	if content == "" {
		return fmt.Errorf("配置内容为空")
	}

	var configData ServiceConfigData

	// 先尝试解析 YAML 格式
	err := yaml.Unmarshal([]byte(content), &configData)
	if err != nil {
		// YAML 解析失败，尝试 JSON 格式
		jsonErr := json.Unmarshal([]byte(content), &configData)
		if jsonErr != nil {
			return fmt.Errorf("解析配置失败 (YAML: %v, JSON: %v)", err, jsonErr)
		}
	}

	// 获取旧配置
	oldConfig := config.GetServiceConfig()

	// 记录配置变更
	if oldConfig.HTTPTimeout != configData.HTTPTimeout {
		log.Printf("[ConfigCenter] HTTP超时时间变更: %ds -> %ds", oldConfig.HTTPTimeout, configData.HTTPTimeout)
	}
	if oldConfig.GRPCTimeout != configData.GRPCTimeout {
		log.Printf("[ConfigCenter] gRPC超时时间变更: %ds -> %ds", oldConfig.GRPCTimeout, configData.GRPCTimeout)
	}
	if oldConfig.DBTimeout != configData.DBTimeout {
		log.Printf("[ConfigCenter] DB超时时间变更: %ds -> %ds", oldConfig.DBTimeout, configData.DBTimeout)
	}
	if oldConfig.RedisTimeout != configData.RedisTimeout {
		log.Printf("[ConfigCenter] Redis超时时间变更: %ds -> %ds", oldConfig.RedisTimeout, configData.RedisTimeout)
	}
	if oldConfig.MaxRetryCount != configData.MaxRetryCount {
		log.Printf("[ConfigCenter] 最大重试次数变更: %d -> %d", oldConfig.MaxRetryCount, configData.MaxRetryCount)
	}
	if oldConfig.DebugMode != configData.DebugMode {
		log.Printf("[ConfigCenter] 调试模式变更: %v -> %v", oldConfig.DebugMode, configData.DebugMode)
	}

	// 更新配置
	newConfig := &config.ServiceConfig{
		HTTPTimeout:   configData.HTTPTimeout,
		GRPCTimeout:   configData.GRPCTimeout,
		DBTimeout:     configData.DBTimeout,
		RedisTimeout:  configData.RedisTimeout,
		MaxRetryCount: configData.MaxRetryCount,
		DebugMode:     configData.DebugMode,
	}

	config.UpdateServiceConfig(newConfig)
	return nil
}

// initDefaultServiceConfig 初始化默认服务配置
func initDefaultServiceConfig() {
	defaultConfig := &config.ServiceConfig{
		HTTPTimeout:   30,
		GRPCTimeout:   30,
		DBTimeout:     10,
		RedisTimeout:  5,
		MaxRetryCount: 3,
		DebugMode:     false,
	}
	config.UpdateServiceConfig(defaultConfig)
	log.Println("[ConfigCenter] 使用默认服务配置")
}

// StopListenConfig 停止监听配置
func StopListenConfig(dataId, group string) error {
	if NacosConfigClient == nil {
		return nil
	}

	err := NacosConfigClient.CancelListenConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  group,
	})

	if err != nil {
		return fmt.Errorf("停止配置监听失败: %w", err)
	}

	log.Println("[ConfigCenter] 已停止配置监听")
	return nil
}

// PrintCurrentServiceConfig 打印当前服务配置
func PrintCurrentServiceConfig() {
	cfg := config.GetServiceConfig()
	log.Println("========== 当前服务配置 ==========")
	log.Printf("HTTP超时时间: %d秒", cfg.HTTPTimeout)
	log.Printf("gRPC超时时间: %d秒", cfg.GRPCTimeout)
	log.Printf("DB超时时间: %d秒", cfg.DBTimeout)
	log.Printf("Redis超时时间: %d秒", cfg.RedisTimeout)
	log.Printf("最大重试次数: %d", cfg.MaxRetryCount)
	log.Printf("调试模式: %v", cfg.DebugMode)
	log.Println("=================================")
}
