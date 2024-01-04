package handler

import (
	"encoding/json"

	"example/vo"

	"github.com/gin-gonic/gin"
	feign "github.com/siawase7179/go_eureka_fegin/eureka/fegin"
	"github.com/sirupsen/logrus"
)

func TokenHandler(ctx *gin.Context) {
	var resByte []byte

	request := ctx.Request
	clientId := request.Header.Get("X-Client-Id")
	clientPassword := request.Header.Get("X-Client-Password")

	logrus.Info("clientId:" + clientId + ", clientPassword:" + clientPassword)

	accountInfo, err := json.Marshal(vo.AccountInfo{ClientId: clientId, ClientPassword: clientPassword})
	if err != nil {
		errorHandler(ctx, 500, vo.AuhResponse{Code: "99999", Result: "Server Error"})
	}

	header := map[string]string{
		"Content-Type": "application/json",
	}

	feignResponse, err := feign.Request("AUTH-SERVER", feign.RequeustOption{
		Path:   "/auth",
		Method: "POST",
		Body:   string(accountInfo),
		Header: header,
	})

	if err != nil {
		errorHandler(ctx, 500, vo.AuhResponse{Code: "99999", Result: "Server Error"})
	}

	logrus.Info("status:" + feignResponse.Response.Status + "body:" + string(feignResponse.Body))

	if feignResponse.Response.StatusCode != 200 {
		errorHandler(ctx, 500, vo.AuhResponse{Code: "99999", Result: "Server Error"})
	}

	var token vo.TokenResponse
	err = json.Unmarshal(feignResponse.Body, &token)
	if err != nil {
		logrus.Error(err.Error())
		errorHandler(ctx, 500, vo.AuhResponse{Code: "99999", Result: "Server Error"})
	} else {
		resByte, err = json.Marshal(token)
		if err != nil {
			logrus.Error(err.Error())
			errorHandler(ctx, 500, vo.AuhResponse{Code: "99999", Result: "Server Error"})
		}
	}

	ctx.Writer.WriteString(string(resByte))
}

func errorHandler(ctx *gin.Context, statusCode int, response vo.AuhResponse) {
	ctx.AbortWithStatusJSON(
		statusCode,
		response,
	)
	return
}
