package dao

import (
	"errors"
	"github.com/inscription-c/cins/inscription/index/tables"
	"gorm.io/gorm"
)

// BlockHeight retrieves the height of the last block in the database.
func (d *DB) BlockHeight() (height uint32, err error) {
	block := &tables.BlockInfo{}
	err = d.DB.Last(block).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	height = block.Height
	return
}
