package handler

import (
	"api-gateway/internal/service"
	"api-gateway/pkg/e"
	"api-gateway/pkg/res"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
)

func UserRegister(ginCtx *gin.Context) {
	var userReq service.UserRequest
	PanicIfUserError(ginCtx.Bind(&userReq))

	// 封装参数进行 gRPC 请求
	userService := ginCtx.Keys["user"].(service.UserServiceClient)
	userResp, err := userService.UserRegister(context.Background(), &userReq)

	PanicIfUserError(err)
	r := res.Response{
		Data:   userResp,
		Status: uint(userResp.Code),
		Msg:    e.GetMsg(uint(userResp.Code)),
	}
	ginCtx.JSON(http.StatusOK, r)
}
