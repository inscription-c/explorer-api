package parser

import "github.com/inscription-c/cins/btcd/rpcclient"

type BLockParserOpts struct {
	client             *rpcclient.Client
	currentBlockNumber int64
}

type OpFunc func(*BLockParserOpts)

func WithClient(client *rpcclient.Client) OpFunc {
	return func(opts *BLockParserOpts) {
		opts.client = client
	}
}

func WithCurrentBlockNumber(blockNumber int64) OpFunc {
	return func(opts *BLockParserOpts) {
		opts.currentBlockNumber = blockNumber
	}
}

type BlockParser struct {
	BLockParserOpts
}

func NewBlockParser(opts ...OpFunc) *BlockParser {
	ops := &BLockParserOpts{}
	for _, opt := range opts {
		opt(ops)
	}

	return &BlockParser{
		BLockParserOpts: *ops,
	}
}

func (b *BlockParser) Run() error {

}
