package handle

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/cins/pkg/util"
	"net/http"
)

type InscribeOrdersResp struct {
	Total int64                `json:"total"`
	List  []*InscribeOrderResp `json:"list"`
}

type InscribeOrderResp struct {
	OrderId        string `json:"order_id"`
	InscriptionId  string `json:"inscription_id"`
	ReceiveAddress string `json:"receive_address"`
	CommitTxId     string `json:"commit_tx_address"`
	RevealTxId     string `json:"reveal_tx_address"`
	RevealTxValue  int64  `json:"reveal_tx_value"`
	Status         int    `json:"status"`
	CreatedAt      int64  `json:"created_at"`
}

func (h *Handler) InscribeOrders(ctx *gin.Context) {
	receiveAddress := ctx.Param("receive_address")
	if receiveAddress == "" {
		ctx.String(http.StatusBadRequest, "receive_address is required")
		return
	}
	page := ctx.Param("page")
	if page == "" {
		page = "1"
	}
	if gconv.Int(page) <= 0 {
		ctx.String(http.StatusBadRequest, "page is invalid")
		return
	}
	if _, err := btcutil.DecodeAddress(receiveAddress, util.ActiveNet.Params); err != nil {
		ctx.String(http.StatusBadRequest, "receive_address is invalid")
		return
	}
	if err := h.doInscribeOrders(ctx, receiveAddress, gconv.Int(page)); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *Handler) doInscribeOrders(ctx *gin.Context, receiveAddress string, page int) error {
	orders, total, err := h.DB().FindInscribeOrdersByReceiveAddress(receiveAddress, page, 10)
	if err != nil {
		return err
	}
	if len(orders) == 0 {
		ctx.Status(http.StatusNotFound)
		return nil
	}

	resp := &InscribeOrdersResp{
		Total: total,
		List:  make([]*InscribeOrderResp, 0, len(orders)),
	}
	for _, order := range orders {
		resp.List = append(resp.List, &InscribeOrderResp{
			OrderId:        order.OrderId,
			InscriptionId:  order.InscriptionId.String(),
			ReceiveAddress: order.ReceiveAddress,
			CommitTxId:     order.CommitTxId,
			RevealTxId:     order.RevealTxId,
			RevealTxValue:  order.RevealTxValue,
			Status:         int(order.Status),
			CreatedAt:      order.CreatedAt.Unix(),
		})
	}
	ctx.JSON(http.StatusOK, resp)
	return nil
}
