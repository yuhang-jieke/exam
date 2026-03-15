package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	__ "github.com/yuhang-jieke/exam/srv/proto"
	"github.com/yuhang-jieke/exam/srv/user-server/basic/config"
	"github.com/yuhang-jieke/exam/srv/user-server/basic/inits"
	"github.com/yuhang-jieke/exam/srv/user-server/handler"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 8081, "The server port")
)

// server is used to implement helloworld.GreeterServer.

func main() {
	flag.Parse()

	// 设置运行时端口（用于Consul注册）
	config.RuntimePort = *port

	// 手动初始化（不再使用init自动初始化，以便在flag解析后执行）
	inits.ConfigInit()
	if err := inits.NacosInit(); err != nil {
		log.Printf("Warning: Nacos init failed: %v", err)
	}

	// 启动配置热更新监听
	// 注意：DataId 和 Group 需要与 Nacos 控制台中的配置一致
	serviceConfigDataId := "demo-config"
	serviceConfigGroup := "DEFAULT_GROUP"
	if config.GlobalConf.Nacos.Group != "" {
		serviceConfigGroup = config.GlobalConf.Nacos.Group
	}
	if err := inits.ListenServiceConfig(serviceConfigDataId, serviceConfigGroup); err != nil {
		log.Printf("Warning: Config listen failed: %v", err)
	}

	// 打印当前服务配置
	inits.PrintCurrentServiceConfig()

	inits.MysqlInit()
	inits.RedisInit()

	// 初始化 RabbitMQ
	if err := inits.RabbitMQInit(); err != nil {
		log.Printf("Warning: RabbitMQ init failed: %v", err)
	} else {
		// 声明商品队列
		if err := inits.DeclareQueue("goods.queue", true, false, false, false); err != nil {
			log.Printf("Warning: Failed to declare goods.queue: %v", err)
		}
	}

	inits.ConsulInit()

	// 启动gRPC服务器
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	__.RegisterEcommerceServiceServer(s, &handler.Server{})
	log.Printf("gRPC server listening at %v", lis.Addr())

	// 启动服务器（在goroutine中）
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// 等待信号并优雅退出
	waitForShutdown(s)
}

// waitForShutdown 等待信号并优雅退出
func waitForShutdown(server *grpc.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-sigChan
	log.Printf("[服务器] 接收到信号: %v，开始优雅退出...", sig)

	// 设置服务为不健康状态（停止心跳）
	if config.ConsulClient != nil {
		config.ConsulClient.SetNotReady()
		log.Println("[服务器] 服务已标记为未就绪")
	}

	// 创建超时上下文，给现有请求5秒时间完成
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 优雅关闭gRPC服务器
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	select {
	case <-ctx.Done():
		log.Println("[服务器] 优雅退出超时，强制停止")
		server.Stop()
	case <-done:
		log.Println("[服务器] gRPC服务器已优雅停止")
	}

	// 从Consul注销服务
	if config.ConsulClient != nil {
		log.Println("[服务器] 正在从Consul注销...")
		if err := config.ConsulClient.Deregister(); err != nil {
			log.Printf("[服务器] 从Consul注销失败: %v", err)
		} else {
			log.Println("[服务器] 已成功从Consul注销")
		}
	}

	log.Println("[服务器] 服务关闭完成")
}
