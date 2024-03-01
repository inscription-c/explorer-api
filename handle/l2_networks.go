package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/explorer-api/constants"
	"net/http"
)

func (h *Handler) L2Networks(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, constants.ActiveChains())
}
