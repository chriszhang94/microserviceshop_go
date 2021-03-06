package routers

import (
	"github.com/gin-gonic/gin"
	"mxshop/goods_web/api"
)

func InitGoodsRouter(Router *gin.RouterGroup){
	GoodsRouter := Router.Group("goods")
	{
		GoodsRouter.GET("", api.List)
		GoodsRouter.POST("", api.New)
		GoodsRouter.GET("/:id", api.GoodsDetail)
		GoodsRouter.DELETE("/:id", api.Delete)
		GoodsRouter.GET("/:id/stocks", api.Stocks)
		GoodsRouter.PUT("/:id", api.Update)
		GoodsRouter.PATCH("/:id", api.UpdateStatus)
	}
}
