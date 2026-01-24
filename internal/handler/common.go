package handler

import (
	"nurture/internal/logic"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type CommonHandler struct{}

func NewCommonHandler() *CommonHandler {
	return &CommonHandler{}
}

func (h *CommonHandler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Response(c, nil, logic.ErrFileRead)
		return
	}
	defer file.Close()

	commonLogic := logic.NewCommonLogic()
	url, err := commonLogic.UploadFile(c, file, header)
	if err != nil {
		response.Response(c, nil, err)
		return
	}
	response.Response(c, url, nil)
}
