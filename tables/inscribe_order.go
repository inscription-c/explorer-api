package tables

import (
	"crypto/md5"
	"fmt"
	"time"
)

type OrderStatus int

const (
	OrderStatusFeeNotEnough OrderStatus = -2
	OrderStatusFail         OrderStatus = -1
	OrderStatusDefault      OrderStatus = 0
	OrderStatusRevealSend   OrderStatus = 1
	OrderStatusSuccess      OrderStatus = 2
)

type InscribeOrder struct {
	Id             uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	OrderId        string `gorm:"column:order_id;type:varchar(255);index:idx_order_id;default:'';NOT NULL"`
	InscriptionId  `gorm:"embedded"`
	RevealAddress  string      `gorm:"column:reveal_address;type:varchar(255);index:idx_reveal_address;default:;NOT NULL"`
	RevealPriKey   string      `gorm:"column:reveal_pri_key;type:varchar(255);default:;NOT NULL"`
	RevealTxId     string      `gorm:"column:reveal_tx_id;type:varchar(255);index:idx_reveal_tx_id;default:;NOT NULL"`
	RevealTxRaw    string      `gorm:"column:reveal_tx_raw;type:mediumtext;default:;NOT NULL"`
	RevealTxValue  int64       `gorm:"column:reveal_tx_value;type:bigint;default:0;NOT NULL"`
	ReceiveAddress string      `gorm:"column:receive_address;type:varchar(255);index:idx_receive_address;default:;NOT NULL"`
	CommitTxId     string      `gorm:"column:commit_tx_id;type:varchar(255);index:idx_commit_tx_id;default:;NOT NULL"`
	Status         OrderStatus `gorm:"column:status;type:int;default:0;NOT NULL"`
	CreatedAt      time.Time   `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt      time.Time   `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (o *InscribeOrder) TableName() string {
	return "inscribe_order"
}

func (o *InscribeOrder) InitOrderId() {
	orderId := fmt.Sprintf("%s%s%d", o.RevealAddress, o.ReceiveAddress, time.Now().UnixMilli())
	o.OrderId = fmt.Sprintf("%x", md5.Sum([]byte(orderId)))
}
