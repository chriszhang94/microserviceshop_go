package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mxshop/order_web/forms"
	"mxshop/order_web/global"
	"mxshop/order_web/proto"
	"net/http"
	"strconv"
)
func Test(ctx *gin.Context){
	pn := ctx.DefaultQuery("pn", "0")
	pnInt, _ := strconv.Atoi(pn)
	pSize := ctx.DefaultQuery("psize", "10")
	pSizeInt, _ := strconv.Atoi(pSize)
	rsp, err := global.GoodsSrvClient.BrandList(context.Background(), &proto.BrandFilterRequest{
		Pages: int32(pnInt),
		PagePerNums: int32(pSizeInt),
	})
	if err != nil{
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	result := make([]interface{}, 0)
	response := gin.H{}
	response["total"] = rsp.Total
	for _, value := range rsp.Data {
		reMap := make(map[string]interface{})
		reMap["id"] = value.Id
		reMap["name"] = value.Name
		reMap["logo"] = value.Logo
		result = append(result, reMap)
	}
	response["data"] = result
	ctx.JSON(http.StatusOK, response)
}
func ShoppingCartList(ctx *gin.Context){
	userId, _ := ctx.Get("userId")
	orderSrvClient := global.OrderSrvClient
	goodsSrvClient := global.GoodsSrvClient
	rsp, err := orderSrvClient.CartItemList(context.Background(), &proto.UserInfo{Id: int32(userId.(uint))})
	if err != nil {
		zap.S().Errorw("[ShoppingCart List] Failed")
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ids := make([]int32, 0)
	for _, item := range rsp.Data{
		ids = append(ids, item.GoodsId)
	}
	if len(ids) == 0{
		ctx.JSON(http.StatusOK, gin.H{
			"total": 0,
		})
		return
	}
	goodsRsp, err := goodsSrvClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{
		Id: ids,
	})
	if err != nil {
		zap.S().Errorw("[List] 批量查询【商品列表】失败")
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	reMap := gin.H{
		"total": goodsRsp.Total,
	}
	goodsList := make([]interface{}, 0)
	for _, item := range rsp.Data{
		for _ , good := range goodsRsp.Data{
			if good.Id == item.GoodsId{
				tmpMap := map[string]interface{}{}
				tmpMap["id"] = item.Id
				tmpMap["goods_id"] = item.GoodsId
				tmpMap["good_name"] = good.Name
				tmpMap["good_image"] = good.GoodsFrontImage
				tmpMap["good_price"] = good.ShopPrice
				tmpMap["nums"] = item.Nums
				tmpMap["checked"] = item.Checked

				goodsList = append(goodsList, tmpMap)
			}
		}
	}
	reMap["data"] = goodsList
	ctx.JSON(http.StatusOK, reMap)
}

func ShoppingCartNew(ctx *gin.Context){
	itemForm := forms.ShopCartItemForm{}
	if err := ctx.ShouldBindJSON(&itemForm); err != nil {
		HandleValidatorError(ctx, err)
		return
	}
	_, err := global.GoodsSrvClient.GetGoodsDetail(context.Background(), &proto.GoodInfoRequest{
		Id: itemForm.GoodsId,
	})
	if err != nil {
		zap.S().Errorw("[List] 查询【商品信息】失败")
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	invRsp, err := global.InventorySrvClient.InvDetail(context.Background(), &proto.GoodsInvInfo{
		GoodsId: itemForm.GoodsId,
	})
	if err != nil {
		zap.S().Errorw("[List] 查询【库存信息】失败")
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	if invRsp.Num < itemForm.Nums {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"nums":"Inventory not enough",
		})
		return
	}
	userId, _ := ctx.Get("userId")
	rsp, err := global.OrderSrvClient.CreateCartItem(context.Background(), &proto.CartItemRequest{
		GoodsId: itemForm.GoodsId,
		UserId: int32(userId.(uint)),
		Nums: itemForm.Nums,
	})

	if err != nil {
		zap.S().Errorw("添加到购物车失败")
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"id": rsp.Id,
		"success": true,
	})
}

func UpdateShoppingCart(ctx *gin.Context){
	id := ctx.Param("id")
	i, _ := strconv.Atoi(id)
	itemForm := forms.ShopCartItemUpdateForm{}
	if err := ctx.ShouldBindJSON(&itemForm); err != nil {
		HandleValidatorError(ctx, err)
		return
	}
	userId, _ := ctx.Get("userId")
	request := proto.CartItemRequest{
		UserId:  int32(userId.(uint)),
		GoodsId: int32(i),
		Nums:    itemForm.Num,
		Checked: false,
	}
	if itemForm.Checked != nil {
		request.Checked = *itemForm.Checked
	}
	_, err := global.OrderSrvClient.UpdateCartItem(context.Background(), &request)
	if err!= nil {
		zap.S().Errorw("更新购物车记录失败")
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.Status(http.StatusOK)
}

func DeleteShoppingCart(ctx *gin.Context){
	id := ctx.Param("id")

	i, _ := strconv.Atoi(id)
	userId, _ := ctx.Get("userId")
	_, err := global.OrderSrvClient.DeleteCartItem(context.Background(), &proto.CartItemRequest{
		UserId:  int32(userId.(uint)),
		GoodsId: int32(i),
	})
	if err != nil {
		zap.S().Errorw("删除购物车记录失败")
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.Status(http.StatusOK)
}
