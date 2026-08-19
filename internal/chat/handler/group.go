package handler

import (
	"nurture/internal/chat/dto"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *ChatHandler) CreateGroup(c *gin.Context) {
	cr := middleware.GetBind[dto.CreateChatGroupReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.CreateGroup(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *ChatHandler) JoinGroup(c *gin.Context) {
	uri := middleware.GetBind[dto.ChatGroupIDUri](c)
	userID := jwtx.GetUserID(c)
	err := h.chatLogic.JoinGroup(c.Request.Context(), userID, uri)
	response.Response(c, nil, err)
}

func (h *ChatHandler) LeaveGroup(c *gin.Context) {
	uri := middleware.GetBind[dto.ChatGroupIDUri](c)
	userID := jwtx.GetUserID(c)
	err := h.chatLogic.LeaveGroup(c.Request.Context(), userID, uri)
	response.Response(c, nil, err)
}

func (h *ChatHandler) TransferOwner(c *gin.Context) {
	uri := middleware.GetBind[dto.ChatGroupIDUri](c)
	cr := middleware.GetBind[dto.ChatGroupTransferReq](c)
	userID := jwtx.GetUserID(c)
	err := h.chatLogic.TransferOwner(c.Request.Context(), userID, uri, cr)
	response.Response(c, nil, err)
}

func (h *ChatHandler) DissolveGroup(c *gin.Context) {
	uri := middleware.GetBind[dto.ChatGroupIDUri](c)
	userID := jwtx.GetUserID(c)
	err := h.chatLogic.DissolveGroup(c.Request.Context(), userID, uri)
	response.Response(c, nil, err)
}

func (h *ChatHandler) ListMyGroups(c *gin.Context) {
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.ListMyGroups(c.Request.Context(), userID)
	response.Response(c, resp, err)
}

func (h *ChatHandler) DiscoverGroups(c *gin.Context) {
	q := middleware.GetBind[dto.ChatGroupDiscoverReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.DiscoverGroups(c.Request.Context(), userID, q)
	response.Response(c, resp, err)
}

func (h *ChatHandler) SearchGroups(c *gin.Context) {
	q := middleware.GetBind[dto.ChatGroupSearchReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.SearchGroups(c.Request.Context(), userID, q)
	response.Response(c, resp, err)
}

func (h *ChatHandler) GroupProfile(c *gin.Context) {
	uri := middleware.GetBind[dto.ChatGroupIDUri](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.GetGroupProfile(c.Request.Context(), userID, uri)
	response.Response(c, resp, err)
}

func (h *ChatHandler) ListMembers(c *gin.Context) {
	uri := middleware.GetBind[dto.ChatGroupIDUri](c)
	q := middleware.GetBind[dto.ChatGroupMemberListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.ListMembers(c.Request.Context(), userID, uri, q)
	response.Response(c, resp, err)
}

func (h *ChatHandler) ListMessages(c *gin.Context) {
	uri := middleware.GetBind[dto.ChatGroupIDUri](c)
	q := middleware.GetBind[dto.ChatGroupMessageListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.ListMessages(c.Request.Context(), userID, uri, q)
	response.Response(c, resp, err)
}

func (h *ChatHandler) MarkSeen(c *gin.Context) {
	uri := middleware.GetBind[dto.ChatGroupIDUri](c)
	userID := jwtx.GetUserID(c)
	err := h.chatLogic.MarkGroupSeen(c.Request.Context(), userID, uri.GroupID, 0)
	response.Response(c, nil, err)
}
