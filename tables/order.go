package tables

import "time"

type OrderStatus int

const (
	OrderStatusFail    OrderStatus = -1
	OrderStatusDefault OrderStatus = 0
	OrderStatusSuccess OrderStatus = 1
)

type InscribeOrder struct {
	Id             uint64 `gorm:"column:id;primary_key;AUTO_INCREMENT;NOT NULL"`
	OrderId        string `gorm:"column:order_id;type:varchar(255);index:idx_order_id;default:'';NOT NULL"`
	InscriptionId  `gorm:"embedded"`
	RevealAddress  string      `gorm:"column:reveal_address;type:varchar(255);index:idx_reveal_address;default:;NOT NULL"`
	RevealPriKey   string      `gorm:"column:reveal_pri_key;type:varchar(255);default:;NOT NULL"`
	RevealTxId     string      `gorm:"column:reveal_tx_id;type:varchar(255);index:idx_reveal_tx_id;default:;NOT NULL"`
	RevealTxRaw    string      `gorm:"column:reveal_tx_raw;type:mediumtext;default:;NOT NULL"`
	ReceiveAddress string      `gorm:"column:receive_address;type:varchar(255);index:idx_receive_address;default:;NOT NULL"`
	CommitTxId     string      `gorm:"column:commit_tx_id;type:varchar(255);index:idx_commit_tx_id;default:;NOT NULL"`
	CommitTxRaw    string      `gorm:"column:commit_tx_raw;type:mediumtext;default:;NOT NULL"`
	Status         OrderStatus `gorm:"column:status;type:int;default:0;NOT NULL"`
	CreatedAt      time.Time   `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
	UpdatedAt      time.Time   `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP;NOT NULL"`
}

func (o *InscribeOrder) TableName() string {
	return "inscribe_order"
}
