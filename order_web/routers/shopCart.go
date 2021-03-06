package routers

import (
	"github.com/gin-gonic/gin"
	"mxshop/order_web/middleware"
	"mxshop/order_web/api"
)

func InitShopCart(Router *gin.RouterGroup){
	ShopCartRouter := Router.Group("shopcarts").Use(middleware.JWTAUth())
	{
		ShopCartRouter.GET("", api.ShoppingCartList)
		ShopCartRouter.POST("", api.ShoppingCartNew)
		ShopCartRouter.DELETE("/:id", api.DeleteShoppingCart)
		ShopCartRouter.PATCH("/:id", api.UpdateShoppingCart)
		ShopCartRouter.GET("test", api.Test)
	}
}
