package runner

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/wire"
	"github.com/inscription-c/cins/btcd/rpcclient"
	"github.com/inscription-c/cins/pkg/signal"
	"github.com/inscription-c/cins/pkg/util"
	"github.com/inscription-c/cins/pkg/util/txscript"
	"github.com/inscription-c/explorer-api/dao"
	"github.com/inscription-c/explorer-api/dao/indexer"
	"github.com/inscription-c/explorer-api/log"
	"github.com/inscription-c/explorer-api/tables"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"sync/atomic"
	"time"
)

type Opts struct {
	client    *rpcclient.Client
	height    uint32
	db        *dao.DB
	indexerDB *indexer.DB
}

type OpFunc func(*Opts)

func WithClient(client *rpcclient.Client) OpFunc {
	return func(opts *Opts) {
		opts.client = client
	}
}

func WithStartHeight(height uint32) OpFunc {
	return func(opts *Opts) {
		opts.height = height
	}
}

func WithDB(db *dao.DB) OpFunc {
	return func(opts *Opts) {
		opts.db = db
	}
}

func WithIndexerDB(db *indexer.DB) OpFunc {
	return func(opts *Opts) {
		opts.indexerDB = db
	}
}

type Runner struct {
	Opts
	errgroup.Group
}

func NewRunner(opts ...OpFunc) *Runner {
	ops := &Opts{}
	for _, opt := range opts {
		opt(ops)
	}

	return &Runner{
		Opts: *ops,
	}
}

func (b *Runner) Start() {
	b.BlockParser()
	b.UpdateRevealTx()
}

func (b *Runner) BlockParser() {
	b.Go(func() error {
		for {
			time.Sleep(time.Second * 5)
			select {
			case <-signal.InterruptChannel:
				return nil
			default:
				err := b.indexBlock()
				if errors.Is(err, ErrDetectReorg) {
					log.Log.Error("UpdateIndex", err)
					return err
				}
				var recoverable *ErrRecoverable
				if errors.As(err, &recoverable) {
					if err := handleReorg(b.db, recoverable.Height, recoverable.Depth); err != nil {
						log.Log.Error("handleReorg", err)
					}
					continue
				}
				if err != nil {
					log.Log.Error("indexBlock", err)
				}
			}
		}
	})
}

func (b *Runner) UpdateRevealTx() {
	b.Go(func() error {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		for range ticker.C {
			select {
			case <-signal.InterruptChannel:
				return nil
			default:
				if err := b.processRevealTx(); err != nil {
					log.Log.Errorf("processRevealTx err: %s", err)
					continue
				}
			}
		}
		return nil
	})
}

func (b *Runner) processRevealTx() error {
	rows, err := b.db.Model(&tables.InscribeOrder{}).Where("status=?", tables.OrderStatusRevealSend).Rows()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var order tables.InscribeOrder
		if err := rows.Scan(&order); err != nil {
			return err
		}
		inscriptionId := &tables.InscriptionId{
			TxId:   order.RevealTxId,
			Offset: 0,
		}
		if _, err := b.indexerDB.GetInscriptionById(inscriptionId); err != nil {
			return err
		}
		order.Status = tables.OrderStatusSuccess
		if err := b.db.Save(&order).Error; err != nil {
			return err
		}
		log.Log.Infof("Inscribe Success, order: %s inscriptionId: %s", order.OrderId, inscriptionId)
	}
	return nil
}

func (b *Runner) indexBlock() error {
	blockCount, err := b.db.BlockCount()
	if err != nil {
		return err
	}
	if blockCount > 0 {
		b.height = blockCount
	}

	endHeight, err := b.client.GetBlockCount()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	blockCh, errCh := b.fetchBlockFrom(ctx, uint32(endHeight))
	if blockCh == nil {
		cancel()
		return nil
	}
	defer func() {
		cancel()
		for range blockCh {
		}
	}()

	for {
		select {
		case block, ok := <-blockCh:
			if !ok {
				return nil
			}

			if err := detectReorg(b.client, b.db, block, b.height); err != nil {
				return err
			}

			if err := b.db.Transaction(func(wtx *dao.DB) error {
				for _, tx := range block.Transactions {
					for idx, txOut := range tx.TxOut {
						pkScript, err := txscript.ParsePkScript(txOut.PkScript)
						if errors.Is(err, txscript.ErrUnsupportedScriptType) {
							continue
						}
						if err != nil {
							return err
						}

						revealTxAddress, err := pkScript.Address(util.ActiveNet.Params)
						if err != nil {
							return err
						}
						order, err := b.db.GetInscribeOrdersByRevealAddress(revealTxAddress.String())
						if err != nil {
							return err
						}
						if order.Id == 0 {
							continue
						}

						if txOut.Value < order.RevealTxValue {
							log.Log.Warn("RevealTxValue is not enough", order.OrderId, tx.TxHash().String(), txOut.Value, order.RevealTxValue)
							order.Status = tables.OrderStatusFeeNotEnough
						} else {
							revealTx, err := b.signRevealTx(tx, &order, idx)
							if err != nil {
								return err
							}
							revealTxHash, err := b.client.SendRawTransaction(revealTx, false)
							if err != nil {
								return err
							}
							log.Log.Infof("RevealTxSendSuccess %s %s %d", order.OrderId, revealTxHash.String(), order.Status)

							revealTxBuf := bytes.NewBufferString("")
							if err := revealTx.Serialize(revealTxBuf); err != nil {
								return err
							}
							order.RevealTxId = revealTxHash.String()
							order.RevealTxRaw = hex.EncodeToString(revealTxBuf.Bytes())
							order.Status = tables.OrderStatusRevealSend
						}
						if err := wtx.UpdateInscribeOrderStatus(b.height, &order); err != nil {
							return err
						}
						log.Log.Info("RevealTx", order.OrderId, order.RevealTxId, order.Status)
						break
					}
				}

				if err := wtx.CreateBlockInfo(&tables.BlockParserInfo{
					Height:    b.height,
					BlockHash: block.BlockHash().String(),
				}); err != nil {
					return err
				}
				if err := wtx.DeleteBlockInfo(b.height - 50); err != nil {
					return err
				}
				if err := updateSavePoints(b.client, wtx, b.height); err != nil {
					return err
				}
				return nil
			}); err != nil {
				return err
			}

			b.height++

		case err = <-errCh:
			return err
		case <-signal.InterruptChannel:
			return nil
		}
	}
}

// signRevealTx is a method of the Inscription struct. It is responsible
// for signing the reveal transaction of the Inscription. It sets the previous
// outpoint of the reveal transaction input, calculates the signature hash, and
// signs the reveal transaction input. It returns an error if there is an error in any of the steps.
func (b *Runner) signRevealTx(commitTx *wire.MsgTx, order *tables.InscribeOrder, idx int) (*wire.MsgTx, error) {
	revealTxData, err := hex.DecodeString(order.RevealTxRaw)
	if err != nil {
		return nil, err
	}
	revealTx := &wire.MsgTx{}
	if err := revealTx.Deserialize(bytes.NewReader(revealTxData)); err != nil {
		return nil, err
	}
	commitTxHash := commitTx.TxHash()
	revealTx.TxIn[0].PreviousOutPoint = *wire.NewOutPoint(&commitTxHash, uint32(idx))

	// It creates a new MultiPrevOutFetcher to fetch previous outputs.
	prevFetcher := txscript.NewMultiPrevOutFetcher(map[wire.OutPoint]*wire.TxOut{
		revealTx.TxIn[idx].PreviousOutPoint: {
			Value:    commitTx.TxOut[idx].Value,
			PkScript: commitTx.TxOut[idx].PkScript,
		},
	})

	// It creates new transaction signature hashes using the reveal transaction and the MultiPrevOutFetcher.
	sigHashes := txscript.NewTxSigHashes(revealTx, prevFetcher)

	// It calculates the signature hash for the reveal transaction.
	signHash, err := txscript.CalcTapScriptSignatureHash(sigHashes, txscript.SigHashDefault, revealTx, 0, prevFetcher, txscript.NewBaseTapLeaf(revealTx.TxIn[0].Witness[1]))
	if err != nil {
		return nil, err
	}

	// It signs the signature hash using the private key.
	revealTxPriKeyBytes, err := hex.DecodeString(order.RevealPriKey)
	if err != nil {
		return nil, err
	}
	priKey, _ := btcec.PrivKeyFromBytes(revealTxPriKeyBytes)
	if err != nil {
		return nil, err
	}
	signature, err := schnorr.Sign(priKey, signHash)
	if err != nil {
		return nil, err
	}

	// It serializes the signature and sets it as the witness of the reveal transaction input.
	sig := signature.Serialize()
	revealTx.TxIn[0].Witness[0] = sig
	return revealTx, nil
}

// fetchBlockFrom is a method that fetches blocks from the blockchain, starting from a specified start height and ending at a specified end height.
// It returns a channel that emits the fetched blocks.
// The method returns an error if there is any issue during the fetching process.
func (b *Runner) fetchBlockFrom(ctx context.Context, endHeight uint32) (chan *wire.MsgBlock, chan error) {
	if b.height > endHeight {
		return nil, nil
	}

	current := uint32(16)
	currentGroupNum := uint32(4)
	lastHeightStart := b.height

	errCh := make(chan error, 1)
	blockCh := make(chan *wire.MsgBlock, current*currentGroupNum)
	currentHeightCh := make(chan []uint32, currentGroupNum*2)

	go func() {
		next := b.height
		defer close(currentHeightCh)
		for height := b.height; height <= endHeight; height++ {
			select {
			case <-ctx.Done():
				return
			default:
				if height-next == current-1 || height == endHeight {
					currentHeightCh <- []uint32{next, height}
					next = height + 1
				}
			}
		}
	}()

	go func() {
		defer close(blockCh)
		errWg := &errgroup.Group{}

		for i := uint32(0); i < currentGroupNum; i++ {
			errWg.Go(func() error {
				for heights := range currentHeightCh {
					start := heights[0]
					end := heights[1]

					errWg := &errgroup.Group{}
					blocks := make([]*wire.MsgBlock, current)
					for i := start; i <= end; i++ {
						height := i
						errWg.Go(func() error {
							block, err := b.getBlockWithRetries(height)
							if err != nil {
								return err
							}
							blocks[(height-start)%uint32(len(blocks))] = block
							return nil
						})
					}
					if err := errWg.Wait(); err != nil {
						return err
					}

					for {
						if atomic.LoadUint32(&lastHeightStart) != start {
							time.Sleep(time.Millisecond)
							continue
						}
						break
					}

					for i := 0; i < len(blocks); i++ {
						if blocks[i] == nil {
							break
						}
						blockCh <- blocks[i]
					}
					atomic.StoreUint32(&lastHeightStart, end+1)
				}
				return nil
			})
		}

		if err := errWg.Wait(); err != nil {
			errCh <- err
		}
	}()

	return blockCh, errCh
}

// getBlockWithRetries is a method that fetches a block from the blockchain at a specified height.
// It retries the fetching process if there is any issue, with an exponential backoff.
// The method returns an error if there is any issue during the fetching process.
func (b *Runner) getBlockWithRetries(height uint32) (*wire.MsgBlock, error) {
	errs := -1
	for {
		select {
		case <-signal.InterruptChannel:
			return nil, signal.ErrInterrupted
		default:
			errs++
			if errs > 0 {
				seconds := 1 << errs
				if seconds > 120 {
					err := errors.New("would sleep for more than 120s, giving up")
					log.Log.Error(err)
				}
				time.Sleep(time.Second * time.Duration(seconds))
			}
			// Get the hash of the block at the specified height.
			hash, err := b.client.GetBlockHash(int64(height))
			if err != nil && !errors.Is(err, rpcclient.ErrClientShutdown) {
				log.Log.Warn("GetBlockHash", err)
				continue
			}
			if errors.Is(err, rpcclient.ErrClientShutdown) {
				return nil, signal.ErrInterrupted
			}

			// Get the block with the obtained hash.
			block, err := b.client.GetBlock(hash)
			if err != nil && !errors.Is(err, rpcclient.ErrClientShutdown) {
				log.Log.Warn("GetBlock", err)
				continue
			}
			if errors.Is(err, rpcclient.ErrClientShutdown) {
				return nil, signal.ErrInterrupted
			}
			return block, nil
		}
	}
}
