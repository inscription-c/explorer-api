package handle

import (
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/cins/inscription/server/config"
	"github.com/inscription-c/explorer-api/handle/middlewares"
)

func (h *Handler) InitRouter() {
	h.Engine().Use(gin.Recovery())
	if config.SrvCfg.Server.EnablePProf {
		pprof.Register(h.Engine())
	}
	if config.SrvCfg.Server.Prometheus {
		p := middlewares.NewPrometheus("gin")
		p.Use(h.Engine())
	}

	h.Engine().Use(middlewares.Cors(config.SrvCfg.Origins...))
	if config.SrvCfg.Sentry.Dsn != "" {
		h.Engine().Use(sentrygin.New(sentrygin.Options{
			Repanic: true,
		}))
	}
	h.Engine().Use(middlewares.Logger())
	h.Engine().GET("/home/page/statistics", h.HomePageStatistics)
	h.Engine().POST("/inscriptions", h.Inscriptions)

	r := h.Engine().Group("/r")
	r.GET("/blockheight", h.BlockHeight)

	h.Engine().GET("/l2/networks", h.L2Networks)
	h.Engine().GET("/estimate-smart-fee", h.EstimateSmartFee)
	h.Engine().GET("/order/status/:order_id", h.OrderStatus)
	h.Engine().GET("/inscribe/orders/:receive_address/:page", h.InscribeOrders)
	h.Engine().POST("/inscribe/order/create/c-brc20-deploy", h.CreateCbr20DeployOrder)
}
