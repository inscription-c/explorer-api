package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/explorer-api/constants"
	"github.com/inscription-c/explorer-api/handle/api_code"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
	"net/http"
)

type HomePageStatisticsResp struct {
	Inscriptions string `json:"inscriptions"`
	StoredData   string `json:"stored_data"`
	TotalFees    string `json:"total_fees"`
}

func (h *Handler) HomePageStatistics(ctx *gin.Context) {
	if err := h.doHomePageStatistics(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, api_code.NewResponse(api_code.InternalServerErr, err.Error()))
		return
	}
}

func (h *Handler) doHomePageStatistics(ctx *gin.Context) error {
	resp := &HomePageStatisticsResp{}

	errWg := &errgroup.Group{}
	errWg.Go(func() error {
		total, err := h.IndexerDB().InscriptionsNum()
		if err != nil {
			return err
		}
		resp.Inscriptions = gconv.String(total)
		return nil
	})
	errWg.Go(func() error {
		storedData, err := h.IndexerDB().InscriptionsStoredData()
		if err != nil {
			return err
		}
		resp.StoredData = gconv.String(storedData)
		return nil
	})
	errWg.Go(func() error {
		totalFees, err := h.IndexerDB().InscriptionsTotalFees()
		if err != nil {
			return err
		}
		btc := decimal.NewFromInt(int64(totalFees)).
			Div(decimal.NewFromInt(int64(constants.OneBtc)))
		resp.TotalFees = btc.String()
		return nil
	})
	if err := errWg.Wait(); err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, resp)
	return nil
}
