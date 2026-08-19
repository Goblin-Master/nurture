package handler

import (
	"nurture/internal/chat/dto"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *ChatHandler) CreateGroup(c *gin.Context) {
	var cr dto.CreateChatGroupReq
	if !h.bindJSON(c, &cr) {
		return
	}
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.CreateGroup(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *ChatHandler) JoinGroup(c *gin.Context) {
	var uri dto.ChatGroupIDUri
	if !h.bindURI(c, &uri) {
		return
	}
	userID := jwtx.GetUserID(c)
	err := h.chatLogic.JoinGroup(c.Request.Context(), userID, uri)
	response.Response(c, nil, err)
}

func (h *ChatHandler) LeaveGroup(c *gin.Context) {
	var uri dto.ChatGroupIDUri
	if !h.bindURI(c, &uri) {
		return
	}
	userID := jwtx.GetUserID(c)
	err := h.chatLogic.LeaveGroup(c.Request.Context(), userID, uri)
	response.Response(c, nil, err)
}

func (h *ChatHandler) TransferOwner(c *gin.Context) {
	var uri dto.ChatGroupIDUri
	if !h.bindURI(c, &uri) {
		return
	}
	var cr dto.ChatGroupTransferReq
	if !h.bindJSON(c, &cr) {
		return
	}
	userID := jwtx.GetUserID(c)
	err := h.chatLogic.TransferOwner(c.Request.Context(), userID, uri, cr)
	response.Response(c, nil, err)
}

func (h *ChatHandler) DissolveGroup(c *gin.Context) {
	var uri dto.ChatGroupIDUri
	if !h.bindURI(c, &uri) {
		return
	}
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
	var q dto.ChatGroupDiscoverReq
	if !h.bindQuery(c, &q) {
		return
	}
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.DiscoverGroups(c.Request.Context(), userID, q)
	response.Response(c, resp, err)
}

func (h *ChatHandler) SearchGroups(c *gin.Context) {
	var q dto.ChatGroupSearchReq
	if !h.bindQuery(c, &q) {
		return
	}
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.SearchGroups(c.Request.Context(), userID, q)
	response.Response(c, resp, err)
}

func (h *ChatHandler) GroupProfile(c *gin.Context) {
	var uri dto.ChatGroupIDUri
	if !h.bindURI(c, &uri) {
		return
	}
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.GetGroupProfile(c.Request.Context(), userID, uri)
	response.Response(c, resp, err)
}

func (h *ChatHandler) ListMembers(c *gin.Context) {
	var uri dto.ChatGroupIDUri
	if !h.bindURI(c, &uri) {
		return
	}
	var q dto.ChatGroupMemberListReq
	if !h.bindQuery(c, &q) {
		return
	}
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.ListMembers(c.Request.Context(), userID, uri, q)
	response.Response(c, resp, err)
}

func (h *ChatHandler) ListMessages(c *gin.Context) {
	var uri dto.ChatGroupIDUri
	if !h.bindURI(c, &uri) {
		return
	}
	var q dto.ChatGroupMessageListReq
	if !h.bindQuery(c, &q) {
		return
	}
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.ListMessages(c.Request.Context(), userID, uri, q)
	response.Response(c, resp, err)
}

func (h *ChatHandler) MarkSeen(c *gin.Context) {
	var uri dto.ChatGroupIDUri
	if !h.bindURI(c, &uri) {
		return
	}
	userID := jwtx.GetUserID(c)
	err := h.chatLogic.MarkGroupSeen(c.Request.Context(), userID, uri.GroupID, 0)
	response.Response(c, nil, err)
}
