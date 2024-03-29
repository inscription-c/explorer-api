package handle

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/explorer-api/constants"
	"github.com/inscription-c/explorer-api/dao/indexer"
	"github.com/inscription-c/explorer-api/handle/api_code"
	"github.com/inscription-c/explorer-api/model"
	"github.com/inscription-c/explorer-api/tables"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SearchType string

const (
	SearchTypeUnknown           SearchType = "unknown"
	SearchTypeEmpty             SearchType = "empty"
	SearchTypeInscriptionId     SearchType = "inscription_id"
	SearchTypeInscriptionNumber SearchType = "inscription_number"
	SearchTypeAddress           SearchType = "address"
	SearchTypeTicker            SearchType = "ticker"
)

type InscriptionsReq struct {
	Search          string   `json:"search"`
	Page            int      `json:"page" binding:"omitempty,min=1"`
	Limit           int      `json:"limit" binding:"omitempty,min=1,max=50"`
	Order           string   `json:"order" binding:"omitempty,oneof=newest oldest"`
	Types           []string `json:"types" binding:"omitempty,dive,oneof=image text json html"`
	InscriptionType string   `json:"inscription_type" binding:"omitempty,oneof=c-brc-20"`
	Charms          []string `json:"charms" binding:"omitempty,dive,oneof=cursed"`
}

func (req *InscriptionsReq) Check() error {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 50
	}
	if req.Order == "" {
		req.Order = "newest"
	}
	return nil
}

type InscriptionsResp struct {
	SearchType SearchType          `json:"search_type"`
	Page       int                 `json:"page"`
	Total      int                 `json:"total"`
	List       []*InscriptionEntry `json:"list"`
}

type InscriptionEntry struct {
	InscriptionId     string          `json:"inscription_id"`
	InscriptionNumber int64           `json:"inscription_number"`
	ContentType       string          `json:"content_type"`
	MediaType         string          `json:"media_type"`
	ContentLength     uint32          `json:"content_length"`
	Timestamp         string          `json:"timestamp"`
	OwnerOutput       string          `json:"owner_output"`
	OwnerAddress      string          `json:"owner_address"`
	Sat               string          `json:"sat"`
	CInsDescription   CInsDescription `json:"c_ins_description"`
	ContentProtocol   string          `json:"content_protocol"`
}

type CInsDescription struct {
	Type      string `json:"type"`
	Chain     string `json:"chain"`
	ChainName string `json:"chain_name"`
	Contract  string `json:"contract"`
}

func (h *Handler) Inscriptions(ctx *gin.Context) {
	req := &InscriptionsReq{}
	if err := ctx.BindJSON(req); err != nil {
		ctx.JSON(http.StatusBadRequest, api_code.NewResponse(api_code.InvalidParams, err.Error()))
		return
	}
	if err := req.Check(); err != nil {
		ctx.JSON(http.StatusBadRequest, api_code.NewResponse(api_code.InvalidParams, err.Error()))
		return
	}
	if err := h.doInscriptions(ctx, req); err != nil {
		ctx.JSON(http.StatusInternalServerError, api_code.NewResponse(api_code.InternalServerErr, err.Error()))
		return
	}
}

func (h *Handler) doInscriptions(ctx *gin.Context, req *InscriptionsReq) error {
	mediaTypes := make([]string, 0)
	contentTypes := make([]string, 0)
	for _, v := range req.Types {
		if v == string(constants.ExtensionHtml) {
			contentTypes = append(contentTypes, string(constants.ContentTypeTextHtml), string(constants.ContentTypeTextHtmlUtf8))
			continue
		}
		mediaTypes = append(mediaTypes, v)
	}

	resp := &InscriptionsResp{
		SearchType: SearchTypeUnknown,
		Page:       req.Page,
		List:       make([]*InscriptionEntry, 0),
	}

	searParams := &indexer.FindProtocolsParams{
		Page:            req.Page,
		Limit:           req.Limit,
		Order:           req.Order,
		MediaTypes:      mediaTypes,
		ContentTypes:    contentTypes,
		Charms:          req.Charms,
		InscriptionType: req.InscriptionType,
	}

	req.Search = strings.TrimSpace(req.Search)
	if req.Search == "" {
		resp.SearchType = SearchTypeEmpty
	}
	if req.Search != "" {
		if constants.InscriptionIdRegexp.MatchString(req.Search) {
			insId := tables.StringToInscriptionId(req.Search)
			ins, err := h.IndexerDB().GetInscriptionById(insId)
			if err != nil {
				return err
			}
			if ins.Id == 0 {
				ctx.Status(http.StatusNotFound)
				return nil
			}
			resp.Total = 1
			resp.SearchType = SearchTypeInscriptionId
			resp.List = append(resp.List, insToScanEntry(&ins))
			ctx.JSON(http.StatusOK, resp)
			return nil
		}

		inscriptionNumber, err := strconv.ParseInt(req.Search, 10, 64)
		if err == nil {
			ins, err := h.IndexerDB().GetInscriptionByInscriptionNum(inscriptionNumber)
			if err != nil {
				return err
			}
			if ins.Id == 0 {
				ctx.Status(http.StatusNotFound)
				return nil
			}
			resp.Total = 1
			resp.SearchType = SearchTypeInscriptionNumber
			resp.List = append(resp.List, insToScanEntry(&ins))
			ctx.JSON(http.StatusOK, resp)
			return nil
		}

		if _, err := btcutil.DecodeAddress(req.Search, h.GetChainParams()); err != nil {
			resp.SearchType = SearchTypeTicker
			searParams.Ticker = req.Search
		} else {
			resp.SearchType = SearchTypeAddress
			searParams.Owner = req.Search
		}
	}

	list, total, err := h.IndexerDB().SearchInscriptions(searParams)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		ctx.Status(http.StatusNotFound)
		return nil
	}

	resp.Total = int(total)

	for _, ins := range list {
		resp.List = append(resp.List, insToScanEntry(ins))
	}

	ctx.JSON(http.StatusOK, resp)
	return nil
}

func insToScanEntry(ins *tables.Inscriptions) *InscriptionEntry {
	return &InscriptionEntry{
		InscriptionId:     ins.InscriptionId.String(),
		InscriptionNumber: ins.InscriptionNum,
		ContentType:       ins.ContentType,
		MediaType:         ins.MediaType,
		ContentLength:     ins.ContentSize,
		Timestamp:         time.Unix(ins.Timestamp, 0).UTC().Format(time.RFC3339),
		OwnerOutput:       model.NewOutPoint(ins.TxId, ins.Index).String(),
		OwnerAddress:      ins.Owner,
		Sat:               gconv.String(ins.Sat),
		CInsDescription: CInsDescription{
			Type:      ins.CInsDescription.Type,
			Chain:     ins.CInsDescription.Chain,
			ChainName: constants.Coins[ins.CInsDescription.Chain].ChainName,
			Contract:  ins.CInsDescription.Contract,
		},
		ContentProtocol: ins.ContentProtocol,
	}
}
