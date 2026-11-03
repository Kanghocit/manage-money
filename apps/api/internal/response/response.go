package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type envelope struct {
	Data  any    `json:"data"`
	Error *Error `json:"error"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, envelope{Data: data, Error: nil})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, envelope{Data: data, Error: nil})
}

func Fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, envelope{
		Data: nil,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	})
}
