package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/explorer-api/handle/api_code"
	"github.com/inscription-c/explorer-api/tables"
	"net/http"
)

type OrderStatusResp struct {
	Status        tables.OrderStatus   `json:"status"`
	InscriptionId tables.InscriptionId `json:"inscription_id"`
}

func (h *Handler) OrderStatus(ctx *gin.Context) {
	orderId := ctx.Param("order_id")
	if orderId == "" {
		ctx.JSON(http.StatusBadRequest, api_code.NewResponse(api_code.InvalidParams, "order_id is required"))
		return
	}
	if err := h.doOrderStatus(ctx, orderId); err != nil {
		ctx.JSON(http.StatusInternalServerError, api_code.NewResponse(api_code.InternalServerErr, err.Error()))
		return
	}
}

func (h *Handler) doOrderStatus(ctx *gin.Context, orderId string) error {
	order, err := h.DB().GetInscribeOrderByOrderId(orderId)
	if err != nil {
		return err
	}
	if order.Id == 0 {
		ctx.Status(http.StatusNotFound)
		return nil
	}

	resp := &OrderStatusResp{
		Status:        order.Status,
		InscriptionId: order.InscriptionId,
	}
	ctx.JSON(http.StatusOK, resp)
	return nil
}
