package config

type Mysql struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type Redis struct {
	Host     string
	Port     int
	Password string
	Database int
}

type Nacos struct {
	Addr      string
	Prot      int
	Namespace string
	DataId    string
	Group     string
}

type RabbitMQ struct {
	Host     string
	Port     int
	User     string
	Password string
	VHost    string
}

type Pay struct {
	PrivateKey string `mapstructure:"private_key"`
	AppId      string `mapstructure:"app_id"`
	NotifyURL  string `mapstructure:"notify_url"`
	ReturnURL  string `mapstructure:"return_url"`
}
type AliPay struct {
	PrivateKey string `mapstructure:"PrivateKey"`
	AppId      string `mapstructure:"AppId"`
	NotifyURL  string `mapstructure:"NotifyURL"`
	ReturnURL  string `mapstructure:"ReturnURL"`
}
type Consul struct {
	Address         string            `mapstructure:"address"`
	Token           string            `mapstructure:"token"`
	Scheme          string            `mapstructure:"scheme"`
	ServiceName     string            `mapstructure:"service_name"`
	ServiceID       string            `mapstructure:"service_id"`
	ServicePort     int               `mapstructure:"service_port"`
	ServiceAddress  string            `mapstructure:"service_addr"`
	TTL             string            `mapstructure:"ttl"`
	CheckTimeout    string            `mapstructure:"check_timeout"`
	DeregisterAfter string            `mapstructure:"deregister_after"`
	Tags            []string          `mapstructure:"tags"`
	Meta            map[string]string `mapstructure:"meta"`
}

// ServiceConfig 服务运行时配置（支持热更新）
type ServiceConfig struct {
	// HTTP请求超时时间(秒)
	HTTPTimeout int `json:"http_timeout"`
	// gRPC请求超时时间(秒)
	GRPCTimeout int `json:"grpc_timeout"`
	// 数据库查询超时时间(秒)
	DBTimeout int `json:"db_timeout"`
	// Redis操作超时时间(秒)
	RedisTimeout int `json:"redis_timeout"`
	// 最大重试次数
	MaxRetryCount int `json:"max_retry_count"`
	// 是否开启调试模式
	DebugMode bool `json:"debug_mode"`
}

type AppConfig struct {
	Mysql
	Redis
	Nacos
	Consul
	RabbitMQ
	AliPay                // 使用 AliPay 字段名匹配配置文件
	Service ServiceConfig `json:"service"`
}
