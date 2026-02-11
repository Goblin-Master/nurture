package handler

import (
	"nurture/internal/dto"
	"nurture/internal/global"
	"nurture/internal/logic"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	postLogic *logic.PostLogic
}

func NewPostHandler() *PostHandler {
	return &PostHandler{
		postLogic: logic.NewPostLogic(),
	}
}

func (h *PostHandler) ListPosts(c *gin.Context) {
	cr := middleware.GetBind[dto.PostListReq](c)
	global.Log.Info(cr)
	resp, err := h.postLogic.List(c.Request.Context(), cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *PostHandler) GetDetail(c *gin.Context) {
	cr := middleware.GetBind[dto.PostDetailReq](c)
	global.Log.Info(cr)
	resp, err := h.postLogic.Detail(c.Request.Context(), cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *PostHandler) NewPost(c *gin.Context) {
	cr := middleware.GetBind[dto.CreatePostReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.NewPost(c.Request.Context(), userID, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *PostHandler) Publish(c *gin.Context) {
	cr := middleware.GetBind[dto.PublishPostReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.Publish(c.Request.Context(), userID, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}
