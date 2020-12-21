package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mxshop/goods_web/global"
	"mxshop/goods_web/proto"
	"net/http"
)


func List(ctx *gin.Context){
	request := &proto.GoodsFilterRequest{}
	goodsSrvClient := global.GoodsSrvClient
	list, err := goodsSrvClient.GoodsList(context.Background(), request)
	if err != nil {
		zap.S().Errorw("[List] Failed to query Goods list")
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	res := gin.H{}
	res["total"] = list.Total
	goodsList := make([]interface{}, 0)
	for _, value := range list.Data {
		goodsList = append(goodsList, map[string]interface{}{
			"id": value.Id,
			"name":        value.Name,
			"goods_brief": value.GoodsBrief,
			"desc":        value.GoodsDesc,
			"ship_free":   value.ShipFree,
			"images":      value.Images,
			"desc_images": value.DescImages,
			"front_image": value.GoodsFrontImage,
			"shop_price":  value.ShopPrice,
			"ctegory": map[string]interface{}{
				"id":   value.Category.Id,
				"name": value.Category.Name,
			},
			"brand": map[string]interface{}{
				"id":   value.Brand.Id,
				"name": value.Brand.Name,
				"logo": value.Brand.Logo,
			},
			"is_hot":  value.IsHot,
			"is_new":  value.IsNew,
			"on_sale": value.OnSale,
		})
	}
	res["data"] = goodsList
	ctx.JSON(http.StatusOK, res)
}