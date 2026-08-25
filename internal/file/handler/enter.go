package handler

import (
	"nurture/internal/file/logic"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FileHandler struct {
	fileLogic logic.IFileLogic
	log       *zap.SugaredLogger
}

func NewFileHandler(fileLogic logic.IFileLogic, log *zap.SugaredLogger) *FileHandler {
	return &FileHandler{
		fileLogic: fileLogic,
		log:       log,
	}
}

func (h *FileHandler) Upload(c *gin.Context) {
	if h.log != nil {
		h.log.Infof("%s: %s", jwtx.GetUserID(c), c.Request.FormValue("file"))
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Response(c, nil, ErrFileRead)
		return
	}
	defer file.Close()

	url, err := h.fileLogic.Upload(c.Request.Context(), file, header)
	response.Response(c, url, err)
}
