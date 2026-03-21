package httpx

import "github.com/gin-gonic/gin"

type Envelope struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func OK(data any) Envelope {
	return Envelope{Code: "OK", Message: "成功", Data: data}
}

func Error(code string, message string) Envelope {
	return Envelope{Code: code, Message: message, Data: nil}
}

func JSON(c *gin.Context, status int, code string, message string, data any) {
	c.JSON(status, Envelope{Code: code, Message: message, Data: data})
}
