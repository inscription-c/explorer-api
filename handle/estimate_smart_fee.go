package handle

import (
	"errors"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/explorer-api/constants"
	"github.com/inscription-c/explorer-api/handle/api_code"
	"golang.org/x/sync/errgroup"
	"net/http"
)

func (h *Handler) EstimateSmartFee(ctx *gin.Context) {
	if err := h.doEstimateSmartFee(ctx); err != nil {
		ctx.JSON(http.StatusInternalServerError, api_code.NewResponse(api_code.InternalServerErr, err.Error()))
		return
	}
}

func (h *Handler) doEstimateSmartFee(ctx *gin.Context) error {
	result := map[string]uint64{}
	errWg := &errgroup.Group{}
	errWg.Go(func() error {
		resp, err := h.RpcClient().EstimateSmartFee(10, &btcjson.EstimateModeConservative)
		if err != nil {
			return err
		}
		if len(resp.Errors) > 0 {
			return errors.New(gconv.String(resp.Errors))
		}
		result["fast"] = uint64(*resp.FeeRate * float64(constants.OneBtc))
		return nil
	})
	errWg.Go(func() error {
		resp, err := h.RpcClient().EstimateSmartFee(20, &btcjson.EstimateModeConservative)
		if err != nil {
			return err
		}
		if len(resp.Errors) > 0 {
			return errors.New(gconv.String(resp.Errors))
		}
		result["normal"] = uint64(*resp.FeeRate * float64(constants.OneBtc))
		return nil
	})
	errWg.Go(func() error {
		resp, err := h.RpcClient().EstimateSmartFee(30, &btcjson.EstimateModeConservative)
		if err != nil {
			return err
		}
		if len(resp.Errors) > 0 {
			return errors.New(gconv.String(resp.Errors))
		}
		result["slow"] = uint64(*resp.FeeRate * float64(constants.OneBtc))
		return nil
	})
	if err := errWg.Wait(); err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, result)
	return nil
}
