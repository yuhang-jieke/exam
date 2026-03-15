package model

import (
	__ "github.com/yuhang-jieke/exam/srv/proto"
	"gorm.io/gorm"
)

type Goods struct {
	gorm.Model
	Name  string  `gorm:"type:varchar(30);comment:名称"`
	Price float64 `gorm:"type:decimal(10,2);comment:价格"`
	Stock int     `gorm:"type:int(11);comment:库存"`
}

func (g *Goods) GoodsAdd(db *gorm.DB) error {
	return db.Create(&g).Error
}

func (g *Goods) GoodsFind(db *gorm.DB, id int64) error {
	return db.Model(&g).Where("id=?", id).First(&g).Error
}

func (g *Goods) UdateGoods(db *gorm.DB, in *__.OrderReq) error {
	return db.Model(&Goods{}).Where("id=?", in.GoodsId).Update("stock", gorm.Expr("stock-?", in.Num)).Error
}

func (g *Goods) DeleteGoods(db *gorm.DB, id int64) error {
	return db.Where("id=?", id).Delete(&g).Error
}

func (g *Goods) UdateGood(db *gorm.DB, in *__.UpdateGoodsReq) interface{} {
	return db.Model(&Goods{}).Where("id=?", in.Id).Update("price", in.Price).Error
}
