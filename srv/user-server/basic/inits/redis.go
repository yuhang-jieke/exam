package inits

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/yuhang-jieke/exam/srv/user-server/basic/config"
)

var ctx = context.Background()

func RedisInit() {
	// 获取Redis配置（Nacos优先，本地配置备选）
	redisConfig := GetRedisConfigFromNacosOrLocal()

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig["Host"], redisConfig["Port"]),
		Password: redisConfig["Password"].(string),
		DB:       redisConfig["Database"].(int),
	})

	err := rdb.Ping(ctx).Err()
	if err != nil {
		panic("redis连接失败")
	}

	// 存入全局变量
	config.RDB = rdb
	fmt.Println("redis连接成功")
}
