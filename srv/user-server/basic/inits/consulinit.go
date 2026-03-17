package inits

import (
	"log"

	"github.com/yuhang-jieke/exam/registry"
	"github.com/yuhang-jieke/exam/srv/user-server/basic/config"
)

func ConsulInit() {

	cfg := &registry.ConsulConfig{
		Address:         config.GlobalConf.Consul.Address,
		Token:           config.GlobalConf.Consul.Token,
		Scheme:          config.GlobalConf.Consul.Scheme,
		ServiceName:     config.GlobalConf.Consul.ServiceName,
		ServiceID:       config.GlobalConf.Consul.ServiceID,
		ServicePort:     config.GlobalConf.Consul.ServicePort,
		ServiceAddress:  config.GlobalConf.Consul.ServiceAddress,
		TTL:             config.GlobalConf.Consul.TTL,
		CheckTimeout:    config.GlobalConf.Consul.CheckTimeout,
		DeregisterAfter: config.GlobalConf.Consul.DeregisterAfter,
		Tags:            config.GlobalConf.Consul.Tags,
		Meta:            config.GlobalConf.Consul.Meta,
	}

	if cfg.Address == "" {
		cfg.Address = "115.190.57.118:8500"
	}
	if cfg.Scheme == "" {
		cfg.Scheme = "http"
	}
	if cfg.TTL == "" {
		cfg.TTL = "5s"
	}
	if cfg.CheckTimeout == "" {
		cfg.CheckTimeout = "3s"
	}
	if cfg.DeregisterAfter == "" {
		cfg.DeregisterAfter = "30s"
	}

	if config.RuntimePort > 0 {
		cfg.ServicePort = config.RuntimePort

		cfg.ServiceID = ""
	}

	if cfg.ServicePort == 0 {
		cfg.ServicePort = 8081
	}

	client, err := registry.NewClient(cfg)
	if err != nil {
		log.Fatalf("[Consul] 创建客户端失败: %v", err)
	}

	if err := client.Register(); err != nil {
		log.Fatalf("[Consul] 注册服务失败: %v", err)
	}

	config.ConsulClient = client

	log.Printf("[Consul] 服务注册成功: %s (端口: %d)", cfg.ServiceName, cfg.ServicePort)
}
