package middleware

import (
	"nurture/internal/constant"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func Authentication(role jwtx.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		UserID, Role, err := jwtx.ParseToken(c)
		if err != nil {
			c.JSON(401, response.Body{
				Code:    -1,
				Message: err.Error(),
				Data:    nil,
			})
			c.Abort()
			return
		}
		if Role < role {
			c.JSON(403, response.Body{
				Code:    -1,
				Message: jwtx.ErrPermissionDenied.Error(),
				Data:    nil,
			})
			c.Abort()
			return
		}
		//将用户id和角色加入ctx
		c.Set(constant.TOKEN_USER_ID, UserID)
		c.Set(constant.TOKEN_ROLE, Role)
		c.Next()
	}
}
