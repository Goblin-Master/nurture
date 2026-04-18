package middleware

import (
	"nurture/internal/pkg/response"
	"reflect"

	"github.com/gin-gonic/gin"
)

func typeKey[T any]() string {
	var v T
	return "request:" + reflect.TypeOf(v).String()
}

func BindJsonMiddleware[T any](c *gin.Context) {
	var cr T
	err := c.ShouldBindJSON(&cr)
	if err != nil {
		response.Response(c, nil, err)
		c.Abort()
		return
	}
	c.Set(typeKey[T](), cr)
}
func BindQueryMiddleware[T any](c *gin.Context) {
	var cr T
	err := c.ShouldBindQuery(&cr)
	if err != nil {
		response.Response(c, nil, err)
		c.Abort()
		return
	}
	c.Set(typeKey[T](), cr)
}
func BindUriMiddleware[T any](c *gin.Context) {
	var cr T
	err := c.ShouldBindUri(&cr)
	if err != nil {
		response.Response(c, nil, err)
		c.Abort()
		return
	}
	c.Set(typeKey[T](), cr)
}
func GetBind[T any](c *gin.Context) T {
	return c.MustGet(typeKey[T]()).(T)
}
