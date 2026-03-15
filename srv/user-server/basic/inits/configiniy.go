package inits

import (
	"github.com/spf13/viper"
	"github.com/yuhang-jieke/exam/srv/user-server/basic/config"
)

func ConfigInit() {
	viper.SetConfigFile("C:\\Users\\ZhuanZ\\Desktop\\yueyeyue\\exam\\srv\\config.yaml")
	viper.ReadInConfig()
	viper.Unmarshal(&config.GlobalConf)
}
