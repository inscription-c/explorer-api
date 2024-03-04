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
