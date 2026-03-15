package inits

import (
	"reflect"
	"testing"
)

func TestBindQueue(t *testing.T) {
	type args struct {
		queueName    string
		routingKey   string
		exchangeName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := BindQueue(tt.args.queueName, tt.args.routingKey, tt.args.exchangeName); (err != nil) != tt.wantErr {
				t.Errorf("BindQueue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ConfigInit()
		})
	}
}

func TestConsulInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ConsulInit()
		})
	}
}

func TestDeclareExchange(t *testing.T) {
	type args struct {
		exchangeName string
		exchangeType string
		durable      bool
		autoDelete   bool
		internal     bool
		noWait       bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeclareExchange(tt.args.exchangeName, tt.args.exchangeType, tt.args.durable, tt.args.autoDelete, tt.args.internal, tt.args.noWait); (err != nil) != tt.wantErr {
				t.Errorf("DeclareExchange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeclareQueue(t *testing.T) {
	type args struct {
		queueName  string
		durable    bool
		autoDelete bool
		exclusive  bool
		noWait     bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeclareQueue(tt.args.queueName, tt.args.durable, tt.args.autoDelete, tt.args.exclusive, tt.args.noWait); (err != nil) != tt.wantErr {
				t.Errorf("DeclareQueue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetConfigFromNacosOrViper(t *testing.T) {
	type args struct {
		dataId string
		group  string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfigFromNacosOrViper(tt.args.dataId, tt.args.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigFromNacosOrViper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetConfigFromNacosOrViper() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMysqlConfigFromNacosOrLocal(t *testing.T) {
	tests := []struct {
		name string
		want map[string]interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMysqlConfigFromNacosOrLocal(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMysqlConfigFromNacosOrLocal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRedisConfigFromNacosOrLocal(t *testing.T) {
	tests := []struct {
		name string
		want map[string]interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRedisConfigFromNacosOrLocal(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRedisConfigFromNacosOrLocal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListenServiceConfig(t *testing.T) {
	type args struct {
		dataId string
		group  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ListenServiceConfig(tt.args.dataId, tt.args.group); (err != nil) != tt.wantErr {
				t.Errorf("ListenServiceConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMysqlInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MysqlInit()
		})
	}
}

func TestNacosInit(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := NacosInit(); (err != nil) != tt.wantErr {
				t.Errorf("NacosInit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrintCurrentServiceConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PrintCurrentServiceConfig()
		})
	}
}

func TestRabbitMQClose(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RabbitMQClose()
		})
	}
}

func TestRabbitMQInit(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RabbitMQInit(); (err != nil) != tt.wantErr {
				t.Errorf("RabbitMQInit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRedisInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RedisInit()
		})
	}
}

func TestStopListenConfig(t *testing.T) {
	type args struct {
		dataId string
		group  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := StopListenConfig(tt.args.dataId, tt.args.group); (err != nil) != tt.wantErr {
				t.Errorf("StopListenConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_initDefaultServiceConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initDefaultServiceConfig()
		})
	}
}

func Test_parseAndApplyConfig(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseAndApplyConfig(tt.args.content); (err != nil) != tt.wantErr {
				t.Errorf("parseAndApplyConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
