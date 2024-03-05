package dao

import (
	"errors"
	"github.com/inscription-c/explorer-api/tables"
	"gorm.io/gorm"
)

func (d *DB) CreateInscribeOrder(order *tables.InscribeOrder) error {
	return d.Create(order).Error
}

func (d *DB) GetInscribeOrderByOrderId(orderId string) (order tables.InscribeOrder, err error) {
	err = d.Where("order_id = ?", orderId).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func (d *DB) FindInscribeOrdersByReceiveAddress(address string, page, limit int) (orders []*tables.InscribeOrder, total int64, err error) {
	db := d.Model(&tables.InscribeOrder{}).Where("receive_address = ?", address)
	err = db.Count(&total).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	err = db.Order("id desc").Offset(limit * (page - 1)).Limit(limit).Find(&orders).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}
