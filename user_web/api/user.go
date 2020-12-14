package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"mxshop/user_web/global/response"
	"mxshop/user_web/proto"
	"mxshop/user_web/global"
	"mxshop/user_web/forms"
	"net/http"
	"strconv"
	"time"
)

func HandleGrpcErrorToHttp(err error, c *gin.Context) {
	//grpc code to http status code
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"msg": e.Message(),
				})
			case codes.Internal:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "Internal Server Error",
				})
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{
					"msg": "parameter error",
				})
			case codes.Unavailable:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "User service unavailable",
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "Other Error",
				})
			}
			return
		}
	}
}

func GetUserList(ctx *gin.Context) {
	zap.S().Debug("visit user list")
	ip := global.ServerConfig.UserSrvConfig.Host
	port := global.ServerConfig.UserSrvConfig.Port
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", ip, port), grpc.WithInsecure())
	if err != nil {
		zap.S().Errorw("[GetUserList] connect to user service failed", "msg", err.Error())
		return
	}
	userSrvClient := proto.NewUserClient(conn)
	pn := ctx.DefaultQuery("pn", "0")
	pnInt, _ := strconv.Atoi(pn)
	pSize := ctx.DefaultQuery("pSize", "10")
	pSizeInt, _ := strconv.Atoi(pSize)
	pageInfo := &proto.PageInfo{
		Pn:    uint32(pnInt),
		PSize: uint32(pSizeInt),
	}
	rsp, err := userSrvClient.GetUserList(context.Background(), pageInfo)
	if err != nil {
		zap.S().Errorw("[GetUserList] failed")
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	result := make([]interface{}, 0)
	for _, value := range rsp.Data {
		user := response.UserResponse{
			Id:       value.Id,
			NickName: value.NickName,
			Birthday: time.Unix(int64(value.BirthDay), 0).Format("2006-01-02"),
			Role:     value.Role,
			Gender:   value.Gender,
			Mobile:   value.Mobile,
		}
		result = append(result, user)
	}
	ctx.JSON(http.StatusOK, result)

}

func PasswordLogin(ctx *gin.Context){
	passwordLoginForm := forms.PasswordLoginForm{}
	if err := ctx.ShouldBind(&passwordLoginForm);err != nil{
		errs, ok := err.(validator.ValidationErrors)
		if !ok{
			ctx.JSON(http.StatusOK, gin.H{
				"msg": err.Error(),
			})
		}
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": errs.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "ok",
	})

}
