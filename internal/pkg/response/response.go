package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	bizErrors "github.com/ZY0506/gin-scaffold/internal/pkg/errors"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int64       `json:"page"`
	PageSize int64       `json:"page_size"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: bizErrors.Success,
		Msg:  "ok",
		Data: data,
	})
}

func Error(c *gin.Context, httpStatus int, code int, msg string) {
	c.AbortWithStatusJSON(httpStatus, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

// Fail 快捷方法，使用 200 状态码返回业务错误
func Fail(c *gin.Context, code int, msg string) {
	Error(c, http.StatusOK, code, msg)
}

// AbortWithError 用于中间件中终止请求并返回错误
func AbortWithError(c *gin.Context, httpStatus int, err *bizErrors.Error) {
	code := bizErrors.ErrInternal
	msg := "internal server error"
	if err != nil {
		code = err.Code
		msg = err.Msg
	}
	c.AbortWithStatusJSON(httpStatus, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

// Page 快捷方法，使用 200 状态码返回分页数据
func Page(c *gin.Context, list interface{}, total, page, pageSize int64) {
	c.JSON(http.StatusOK, Response{
		Code: bizErrors.Success,
		Msg:  "ok",
		Data: PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}
