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

func (d *DB) GetInscribeOrdersByRevealAddress(address string) (order tables.InscribeOrder, err error) {
	err = d.Where("reveal_address = ? and status = ?", address, tables.OrderStatusDefault).First(&order).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

func (d *DB) UpdateInscribeOrderStatus(height uint32, newOrder *tables.InscribeOrder) error {
	old := &tables.InscribeOrder{}
	err := d.Where("id = ?", newOrder.Id).First(old).Error
	if err != nil {
		return err
	}

	err = d.Save(newOrder).Error
	if err != nil {
		return err
	}
	sql := d.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Save(old)
	})
	return d.AddUndoLog(height, sql)
}
