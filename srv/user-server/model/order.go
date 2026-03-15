package model

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	OrderCode string  `gorm:"type:varchar(32);comment:订单号"`
	GoodsId   int     `gorm:"type:int(11);comment:商品ID"`
	Num       int     `gorm:"type:int(11);comment:数量"`
	Price     float64 `gorm:"type:decimal(10,2);comment:总价格"`
}

func (o *Order) OrderAdd(db *gorm.DB) error {
	return db.Create(&o).Error
}

func (o *Order) FindId(db *gorm.DB, id int64) (Order, error) {
	var list Order
	err := db.Model(&Order{}).Where("goods_id=?", id).First(&list).Error
	return list, err
}
