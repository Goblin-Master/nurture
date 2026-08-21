package dto

type (
	CreateChatGroupReq struct {
		Name        string `json:"name" binding:"required"`
		Avatar      string `json:"avatar"`
		Description string `json:"description"`
		MemberLimit int32  `json:"member_limit" binding:"required"`
	}
	CreateChatGroupResp struct {
		GroupID string `json:"group_id"`
		Message string `json:"message"`
	}
)

type (
	ChatGroupIDUri struct {
		GroupID string `uri:"group_id" binding:"required"`
	}
	ChatGroupTransferReq struct {
		TargetUserID string `json:"target_user_id" binding:"required"`
	}
	ChatGroupListItem struct {
		GroupID               string `json:"group_id"`
		Name                  string `json:"name"`
		Avatar                string `json:"avatar"`
		Description           string `json:"description"`
		MemberLimit           int32  `json:"member_limit"`
		MemberCount           int32  `json:"member_count"`
		Role                  string `json:"role"`
		UnreadCount           int64  `json:"unread_count"`
		Ctime                 int64  `json:"ctime"`
		Utime                 int64  `json:"utime"`
		LastMessageFromUserID string `json:"last_message_from_user_id"`
		LastMessageFromName   string `json:"last_message_from_name"`
		LastMessageType       string `json:"last_message_type"`
		LastMessageContent    string `json:"last_message_content"`
		LastMessageTime       int64  `json:"last_message_time"`
	}
	ListMyChatGroupsResp struct {
		Items []ChatGroupListItem `json:"items"`
	}
)

type (
	ChatGroupDiscoverReq struct {
		Seed   string `form:"seed"`
		Cursor string `form:"cursor"`
		Limit  int    `form:"limit"`
	}
	ChatGroupDiscoverItem struct {
		GroupID     string `json:"group_id"`
		Name        string `json:"name"`
		Avatar      string `json:"avatar"`
		MemberCount int32  `json:"member_count"`
		MemberLimit int32  `json:"member_limit"`
	}
	ChatGroupDiscoverResp struct {
		Seed       string                  `json:"seed"`
		Items      []ChatGroupDiscoverItem `json:"items"`
		HasMore    bool                    `json:"has_more"`
		NextCursor string                  `json:"next_cursor,omitempty"`
	}
)

type (
	ChatGroupSearchReq struct {
		Keyword string `form:"keyword" binding:"required"`
		Limit   int    `form:"limit"`
	}
	ChatGroupSearchResp struct {
		Items []ChatGroupDiscoverItem `json:"items"`
	}
)

type (
	ChatGroupProfileOwner struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
	}
	ChatGroupProfile struct {
		GroupID     string `json:"group_id"`
		Name        string `json:"name"`
		Avatar      string `json:"avatar"`
		Description string `json:"description"`
		MemberCount int32  `json:"member_count"`
		MemberLimit int32  `json:"member_limit"`
		Ctime       int64  `json:"ctime"`
		Utime       int64  `json:"utime"`
	}
	ChatGroupProfileResp struct {
		Group          ChatGroupProfile      `json:"group"`
		Owner          ChatGroupProfileOwner `json:"owner"`
		MembersPreview []ChatGroupMemberItem `json:"members_preview"`
	}
)

type (
	ChatGroupMemberListReq struct {
		Page     int `form:"page"`
		PageSize int `form:"page_size"`
	}
	ChatGroupMemberItem struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Avatar   string `json:"avatar"`
		Role     string `json:"role"`
		JoinTime int64  `json:"join_time"`
	}
	ChatGroupMemberListResp struct {
		Items    []ChatGroupMemberItem `json:"items"`
		Page     int                   `json:"page"`
		PageSize int                   `json:"page_size"`
		HasMore  bool                  `json:"has_more"`
	}
)

type (
	ChatGroupMessageListReq struct {
		Before string `form:"before"`
		After  string `form:"after"`
		Limit  int    `form:"limit"`
	}
	ChatGroupMessageItem struct {
		MessageID  string `json:"message_id"`
		GroupID    string `json:"group_id"`
		FromUserID string `json:"from_user_id"`
		Type       string `json:"type"`
		Content    string `json:"content"`
		Ctime      int64  `json:"ctime"`
	}
	ChatGroupMessageListResp struct {
		Items      []ChatGroupMessageItem `json:"items"`
		HasMore    bool                   `json:"has_more"`
		NextBefore string                 `json:"next_before,omitempty"`
		NextAfter  string                 `json:"next_after,omitempty"`
	}
)

type (
	ChatDirectMessageUserUri struct {
		UserID string `uri:"user_id" binding:"required"`
	}
	ChatDirectMessageListReq struct {
		Before string `form:"before"`
		After  string `form:"after"`
		Limit  int    `form:"limit"`
	}
	ChatDirectMessageItem struct {
		MessageID  string `json:"message_id"`
		FromUserID string `json:"from_user_id"`
		ToUserID   string `json:"to_user_id"`
		Type       string `json:"type"`
		Content    string `json:"content"`
		Ctime      int64  `json:"ctime"`
	}
	ChatDirectMessageListResp struct {
		Items      []ChatDirectMessageItem `json:"items"`
		HasMore    bool                    `json:"has_more"`
		NextBefore string                  `json:"next_before,omitempty"`
		NextAfter  string                  `json:"next_after,omitempty"`
	}
)
