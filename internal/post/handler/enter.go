package handler

import (
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"
	"nurture/internal/post/dto"
	"nurture/internal/post/logic"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	postLogic logic.IPostLogic
}

func NewPostHandler(postLogic logic.IPostLogic) *PostHandler {
	return &PostHandler{
		postLogic: postLogic,
	}
}

func (h *PostHandler) Home(c *gin.Context) {
	cr := middleware.GetBind[dto.PostHomeListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.Home(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *PostHandler) ListByTag(c *gin.Context) {
	cr := middleware.GetBind[dto.PostTagListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.ListByTag(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *PostHandler) Search(c *gin.Context) {
	cr := middleware.GetBind[dto.PostSearchListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.Search(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *PostHandler) GetDetail(c *gin.Context) {
	cr := middleware.GetBind[dto.PostDetailReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.Detail(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *PostHandler) NewPost(c *gin.Context) {
	cr := middleware.GetBind[dto.CreatePostReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.NewPost(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *PostHandler) Publish(c *gin.Context) {
	cr := middleware.GetBind[dto.PublishPostReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.Publish(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *PostHandler) UpdateDraft(c *gin.Context) {
	uri := middleware.GetBind[dto.PostDetailReq](c)
	body := middleware.GetBind[dto.UpdateDraftReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.UpdateDraft(c.Request.Context(), userID, uri, body)
	response.Response(c, resp, err)
}

func (h *PostHandler) DeleteDraft(c *gin.Context) {
	uri := middleware.GetBind[dto.PostDetailReq](c)
	userID := jwtx.GetUserID(c)
	err := h.postLogic.DeleteDraft(c.Request.Context(), userID, uri)
	response.Response(c, nil, err)
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	uri := middleware.GetBind[dto.PostDetailReq](c)
	userID := jwtx.GetUserID(c)
	err := h.postLogic.DeletePost(c.Request.Context(), userID, uri)
	response.Response(c, nil, err)
}

func (h *PostHandler) CreateComment(c *gin.Context) {
	uri := middleware.GetBind[dto.PublishPostReq](c)
	body := middleware.GetBind[dto.CreateCommentReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.CreateComment(c.Request.Context(), userID, uri.PostID, body)
	response.Response(c, resp, err)
}

func (h *PostHandler) ListMyPosts(c *gin.Context) {
	cr := middleware.GetBind[dto.PostMyListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.ListMyPosts(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *PostHandler) Following(c *gin.Context) {
	cr := middleware.GetBind[dto.PostMyListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.Following(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *PostHandler) ListComments(c *gin.Context) {
	uri := middleware.GetBind[dto.PostDetailReq](c)
	query := middleware.GetBind[dto.CommentListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.ListComments(c.Request.Context(), userID, uri.PostID, query)
	response.Response(c, resp, err)
}

func (h *PostHandler) ListReplies(c *gin.Context) {
	uri := middleware.GetBind[dto.CommentRepliesReq](c)
	query := middleware.GetBind[dto.CommentListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.ListReplies(c.Request.Context(), userID, uri, query)
	response.Response(c, resp, err)
}

func (h *PostHandler) DeleteComment(c *gin.Context) {
	uri := middleware.GetBind[dto.CommentDeleteReq](c)
	userID := jwtx.GetUserID(c)
	err := h.postLogic.DeleteComment(c.Request.Context(), userID, uri)
	response.Response(c, nil, err)
}

func (h *PostHandler) UpdateComment(c *gin.Context) {
	uri := middleware.GetBind[dto.CommentUpdateReq](c)
	body := middleware.GetBind[dto.UpdateCommentReq](c)
	userID := jwtx.GetUserID(c)
	err := h.postLogic.UpdateComment(c.Request.Context(), userID, uri, body)
	response.Response(c, nil, err)
}

func (h *PostHandler) LikeComment(c *gin.Context) {
	uri := middleware.GetBind[dto.CommentLikeReq](c)
	userID := jwtx.GetUserID(c)
	err := h.postLogic.LikeComment(c.Request.Context(), userID, uri)
	response.Response(c, nil, err)
}

func (h *PostHandler) UnlikeComment(c *gin.Context) {
	uri := middleware.GetBind[dto.CommentLikeReq](c)
	userID := jwtx.GetUserID(c)
	err := h.postLogic.UnlikeComment(c.Request.Context(), userID, uri)
	response.Response(c, nil, err)
}

func (h *PostHandler) LikePost(c *gin.Context) {
	uri := middleware.GetBind[dto.PostDetailReq](c)
	userID := jwtx.GetUserID(c)
	err := h.postLogic.LikePost(c.Request.Context(), userID, uri)
	response.Response(c, nil, err)
}

func (h *PostHandler) UnlikePost(c *gin.Context) {
	uri := middleware.GetBind[dto.PostDetailReq](c)
	userID := jwtx.GetUserID(c)
	err := h.postLogic.UnlikePost(c.Request.Context(), userID, uri)
	response.Response(c, nil, err)
}

func (h *PostHandler) CollectPost(c *gin.Context) {
	uri := middleware.GetBind[dto.PostDetailReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.CollectPost(c.Request.Context(), userID, uri)
	response.Response(c, resp, err)
}

func (h *PostHandler) UncollectPost(c *gin.Context) {
	uri := middleware.GetBind[dto.PostDetailReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.UncollectPost(c.Request.Context(), userID, uri)
	response.Response(c, resp, err)
}

func (h *PostHandler) ListMyCollections(c *gin.Context) {
	cr := middleware.GetBind[dto.PostMyListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.ListMyCollections(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *PostHandler) ListMyDrafts(c *gin.Context) {
	cr := middleware.GetBind[dto.PostMyListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.ListMyDrafts(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *PostHandler) ListMyMilestones(c *gin.Context) {
	cr := middleware.GetBind[dto.PostMyListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.postLogic.ListMyMilestones(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *PostHandler) AdminCreateTag(c *gin.Context) {
	body := middleware.GetBind[dto.AdminTagCreateReq](c)
	resp, err := h.postLogic.AdminCreateTag(c.Request.Context(), body)
	response.Response(c, resp, err)
}

func (h *PostHandler) AdminDeleteTag(c *gin.Context) {
	uri := middleware.GetBind[dto.AdminTagDeleteUri](c)
	err := h.postLogic.AdminDeleteTag(c.Request.Context(), uri)
	response.Response(c, nil, err)
}

func (h *PostHandler) ListTags(c *gin.Context) {
	query := middleware.GetBind[dto.TagListReq](c)
	resp, err := h.postLogic.ListTags(c.Request.Context(), query)
	response.Response(c, resp, err)
}
