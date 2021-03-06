package api

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"mxshop/user_web/forms"
	"mxshop/user_web/global"
	"mxshop/user_web/global/response"
	"mxshop/user_web/middleware"
	"mxshop/user_web/models"
	"mxshop/user_web/proto"
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

// /user/list
func GetUserList(ctx *gin.Context) {
	userSrvClient := global.UserSrvClient
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

	if !store.Verify(passwordLoginForm.CaptchaId, passwordLoginForm.Captcha, true){
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "Wrong Captcha",
		})
		return
	}

	userSrvClient := global.UserSrvClient

	//login logic
	rsp, err := userSrvClient.GetUserByMobile(context.Background(), &proto.MobileRequest{
		Mobile: passwordLoginForm.Mobile,
	})
	if err != nil {
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	if passrsp, _ := userSrvClient.CheckPassword(context.Background(), &proto.PasswordCheckInfo{
		Password: passwordLoginForm.Password,
		EncryptedPassword: rsp.Password,
	}); !passrsp.Success{
		ctx.JSON(http.StatusOK, gin.H{
			"msg": "Wrong Password",
		})
		return
	}
	j := middleware.NewJWT()
	claims := models.CustomClaims{
		ID: uint(rsp.Id),
		NickName: rsp.NickName,
		AuthorityId: uint(rsp.Role),
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),
			ExpiresAt: time.Now().Unix() + 60 * 60 * 24 * 30,  //7 days expire,
			Issuer: "chris",
		},
	}
	token, err := j.CreateToken(claims)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "JWT Token Error",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"msg": "Success",
		"id": rsp.Id,
		"nickname": rsp.NickName,
		"token": token,
		"expired_at": (time.Now().Unix() + 60 * 60 * 24 * 7)*1000,
	})

}

func Register(ctx *gin.Context){
	registerForm := forms.RegisterForm{}
	if err := ctx.ShouldBind(&registerForm);err != nil{
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

	userSrvClient := global.UserSrvClient
	rsp, err := userSrvClient.CreateUser(context.Background(), &proto.CreateUserInfo{
		NickName: registerForm.Nickname,
		Password: registerForm.Password,
		Mobile: registerForm.Mobile,
	});
	if err != nil{
		ctx.JSON(http.StatusOK, gin.H{
			"msg": "Register Failed",
			"detail": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id": rsp.Id,
		"msg": "Register Succeed",
	})

}
