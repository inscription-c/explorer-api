package handle

import (
	"github.com/btcsuite/btcd/btcjson"
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/explorer-api/constants"
	"golang.org/x/sync/errgroup"
	"net/http"
)

func (h *Handler) EstimateSmartFee(ctx *gin.Context) {
	if err := h.doEstimateSmartFee(ctx); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *Handler) doEstimateSmartFee(ctx *gin.Context) error {
	result := map[string]uint32{}
	errWg := &errgroup.Group{}
	errWg.Go(func() error {
		resp, err := h.RpcClient().EstimateSmartFee(10, &btcjson.EstimateModeConservative)
		if err != nil {
			return err
		}
		result["fast"] = uint32(*resp.FeeRate * float64(constants.OneBtc))
		return nil
	})
	errWg.Go(func() error {
		resp, err := h.RpcClient().EstimateSmartFee(20, &btcjson.EstimateModeConservative)
		if err != nil {
			return err
		}
		result["normal"] = uint32(*resp.FeeRate * float64(constants.OneBtc))
		return nil
	})
	errWg.Go(func() error {
		resp, err := h.RpcClient().EstimateSmartFee(30, &btcjson.EstimateModeConservative)
		if err != nil {
			return err
		}
		result["slow"] = uint32(*resp.FeeRate * float64(constants.OneBtc))
		return nil
	})
	if err := errWg.Wait(); err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, result)
	return nil
}
