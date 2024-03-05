package dao

import (
	"errors"
	"github.com/inscription-c/explorer-api/tables"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BlockCount retrieves the total number of blocks in the database.
// It returns the count as an uint32 and any error encountered.
func (d *DB) BlockCount() (count uint32, err error) {
	block := &tables.BlockParserInfo{}
	// Retrieve the last block
	err = d.DB.Order("id desc").First(block).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
		return
	}
	count = block.Height + 1
	return
}

func (d *DB) BlockHash(height ...uint32) (blockHash string, err error) {
	blockInfo := &tables.BlockParserInfo{}
	if len(height) == 0 {
		if err = d.Last(blockInfo).Error; err != nil {
			return
		}
	} else {
		if err = d.Where("block_number = ?", height[0]).First(blockInfo).Error; err != nil {
			return
		}
	}
	blockHash = blockInfo.BlockHash
	return
}

func (d *DB) DeleteBlockInfo(height uint32) error {
	deleted := make([]*tables.BlockParserInfo, 0)
	err := d.Clauses(clause.Returning{}).Where("block_number < ?", height).Delete(&deleted).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if len(deleted) == 0 {
		return nil
	}
	sql := d.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Create(&deleted)
	})
	return d.AddUndoLog(height, sql)
}

func (d *DB) CreateBlockInfo(block *tables.BlockParserInfo) error {
	err := d.Create(block).Error
	if err != nil {
		return err
	}
	sql := d.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Delete(block)
	})
	return d.AddUndoLog(block.Height, sql)
}
