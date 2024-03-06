package handle

import (
	"bytes"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/cins/constants"
	"github.com/inscription-c/cins/inscription"
	"github.com/inscription-c/cins/inscription/index/tables"
	"github.com/inscription-c/cins/pkg/util"
	constants2 "github.com/inscription-c/explorer-api/constants"
	"github.com/inscription-c/explorer-api/handle/api_code"
	tables2 "github.com/inscription-c/explorer-api/tables"
	"net/http"
)

type CreateCbr20DeployOrderReq struct {
	Postage        int64  `json:"postage" binding:"min=330,max=10000"`
	FeatRate       int64  `json:"fee_rate" binding:"gt=0"`
	Ticker         string `json:"ticker" binding:"required"`
	TotalSupply    string `json:"total_supply" binding:"required"`
	LimitPerMint   string `json:"limit_per_mint" binding:"required"`
	L2NetWork      string `json:"l2_network" binding:"required"`
	Contract       string `json:"contract" binding:"required"`
	ReceiveAddress string `json:"receive_address" binding:"required"`
}

func (h *Handler) CreateCbr20DeployOrder(ctx *gin.Context) {
	req := &CreateCbr20DeployOrderReq{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, api_code.NewResponse(api_code.InvalidParams, err.Error()))
		return
	}
	if constants2.ChainName(req.L2NetWork) == "" {
		ctx.JSON(http.StatusBadRequest, api_code.NewResponse(api_code.InvalidParams, "invalid l2_network"))
		return
	}
	if err := h.doCreateCbr20DeployOrder(ctx, req); err != nil {
		ctx.JSON(http.StatusBadRequest, api_code.NewResponse(api_code.InternalServerErr, err.Error()))
		return
	}
}

func (h *Handler) doCreateCbr20DeployOrder(ctx *gin.Context, req *CreateCbr20DeployOrderReq) error {
	priKey, err := btcec.NewPrivateKey()
	if err != nil {
		return err
	}
	internalKey := priKey.PubKey()

	cbrc20 := &util.CBRC20{
		Protocol:  constants.ProtocolCBRC20,
		Operation: constants.OperationDeploy,
		Tick:      req.Ticker,
		Max:       req.TotalSupply,
		Limit:     req.LimitPerMint,
		Decimals:  constants.DecimalsDefault,
	}

	revealScript, err := inscription.InscriptionToScript(
		internalKey,
		inscription.Header{
			CInsDescription: &tables.CInsDescription{
				Type:     constants.CInsDescriptionTypeBlockchain,
				Chain:    req.L2NetWork,
				Contract: req.Contract,
			},
			ContentType: constants.ContentTypeJson,
		},
		cbrc20,
	)
	if err != nil {
		return err
	}

	// Generate the script address
	controlBlock, taprootAddress, err := inscription.RevealScriptAddress(internalKey, revealScript)
	if err != nil {
		return err
	}

	// Create the witness for the transaction
	revealTxWitness := make([][]byte, 0)
	revealTxWitness = append(revealTxWitness, make([]byte, 64))
	revealTxWitness = append(revealTxWitness, revealScript)
	controlBlockBytes, err := controlBlock.ToBytes()
	if err != nil {
		return err
	}
	revealTxWitness = append(revealTxWitness, controlBlockBytes)
	taprootScript, err := txscript.PayToAddrScript(taprootAddress)
	if err != nil {
		return err
	}

	// Create the transaction input
	revealTxIn := &wire.TxIn{
		SignatureScript: taprootScript,
		Witness:         revealTxWitness,
		Sequence:        0xFFFFFFFD,
	}

	// Create the transaction output
	destAddrScript, err := util.AddressScript(req.ReceiveAddress, util.ActiveNet.Params)
	if err != nil {
		return err
	}
	revealTxOutput := wire.NewTxOut(req.Postage, destAddrScript)

	// Create the reveal transaction
	revealTx := wire.NewMsgTx(2)
	revealTx.AddTxIn(revealTxIn)
	revealTx.AddTxOut(revealTxOutput)

	revealTxRaw := bytes.NewBufferString("")
	if err := revealTx.Serialize(revealTxRaw); err != nil {
		return err
	}

	txFee := inscription.CalculateTxFee(revealTx, req.FeatRate)
	revealTxValue := txFee + req.Postage

	order := &tables2.InscribeOrder{
		RevealAddress:  taprootAddress.String(),
		RevealPriKey:   hex.EncodeToString(priKey.Serialize()),
		RevealTxRaw:    hex.EncodeToString(revealTxRaw.Bytes()),
		RevealTxValue:  revealTxValue,
		ReceiveAddress: req.ReceiveAddress,
	}
	order.InitOrderId()
	if err := h.DB().CreateInscribeOrder(order); err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, gin.H{
		"order_id": order.OrderId,
		"address":  taprootAddress.String(),
		"value":    revealTxValue,
	})
	return nil
}
