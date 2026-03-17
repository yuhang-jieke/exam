package inits

import (
	"fmt"

	"github.com/yuhang-jieke/exam/srv/user-server/basic/config"
	"github.com/yuhang-jieke/exam/srv/user-server/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func MysqlInit() {
	var err error

	mysqlConfig := GetMysqlConfigFromNacosOrLocal()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlConfig["User"], mysqlConfig["Password"], mysqlConfig["Host"], mysqlConfig["Port"], mysqlConfig["Database"])
	config.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败")
	}
	fmt.Println("数据库连接成功")
	err = config.DB.AutoMigrate(&model.Goods{}, &model.Order{})
	if err != nil {
		panic("数据表迁移失败")
	}
	fmt.Println("数据表迁移成功")
}
