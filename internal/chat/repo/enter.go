package repo

import (
	"context"
	"fmt"
	"nurture/internal/chat/repo/dao"
	"nurture/internal/pkg/zapx"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ChatGroupListItem struct {
	GroupID               string
	Name                  string
	Avatar                string
	Description           string
	MemberLimit           int32
	MemberCount           int32
	Role                  string
	UnreadCount           int64
	Ctime                 int64
	Utime                 int64
	LastMessageFromUserID string
	LastMessageFromName   string
	LastMessageType       string
	LastMessageContent    string
	LastMessageTime       int64
}

type ChatGroupMemberProfile struct {
	UserID   string
	Username string
	Avatar   string
	Role     string
	JoinTime int64
}

type ChatGroupMessageItem struct {
	MessageID  string
	GroupID    string
	FromUserID string
	Type       string
	Content    string
	Ctime      int64
}

type ChatDirectMessageItem struct {
	MessageID  string
	FromUserID string
	ToUserID   string
	Type       string
	Content    string
	Ctime      int64
}

type ChatOutboxEvent struct {
	ID         int64
	EventID    string
	RoutingKey string
	Payload    string
	Attempts   int32
	Ctime      int64
}

type ChatGroupDiscoverItem struct {
	GroupID     string
	Name        string
	Avatar      string
	MemberCount int32
	MemberLimit int32
	SortKey     string
}

type ChatGroupProfileItem struct {
	GroupID     string
	Name        string
	Avatar      string
	Description string
	MemberCount int32
	MemberLimit int32
	Ctime       int64
	Utime       int64
	OwnerUserID string
	OwnerName   string
	OwnerAvatar string
}

type IChatRepo interface {
	CreateGroup(ctx context.Context, groupID, ownerID, name, avatar, description string, memberLimit int32, now int64) error
	JoinGroup(ctx context.Context, groupID, userID string, now int64) error
	LeaveGroup(ctx context.Context, groupID, userID string, now int64) error
	TransferOwner(ctx context.Context, groupID, ownerID, targetUserID string, now int64) error
	DissolveGroup(ctx context.Context, groupID, ownerID string, now int64) error
	UpdateMemberLastSeenTime(ctx context.Context, groupID, userID string, lastSeenTime int64) error

	ListMyGroups(ctx context.Context, userID string) ([]ChatGroupListItem, error)
	ListDiscoverGroups(ctx context.Context, userID, seed, cursorSortKey, cursorGroupID string, limit int) ([]ChatGroupDiscoverItem, string, bool, error)
	SearchGroupsByName(ctx context.Context, keyword string, limit int) ([]ChatGroupDiscoverItem, error)
	GetGroupProfile(ctx context.Context, groupID string, memberPreviewLimit int) (ChatGroupProfileItem, []ChatGroupMemberProfile, error)
	ListMembersWithProfile(ctx context.Context, groupID string, page, pageSize int) ([]ChatGroupMemberProfile, bool, error)

	SaveMessage(ctx context.Context, groupID, messageID, fromUserID, msgType, content string, now int64, outbox ChatOutboxEvent) (bool, error)
	SaveDirectMessage(ctx context.Context, messageID, fromUserID, toUserID, msgType, content string, now int64, outbox ChatOutboxEvent) (bool, error)
	ListMessagesLatest(ctx context.Context, groupID string, limit int) ([]ChatGroupMessageItem, error)
	ListMessagesBefore(ctx context.Context, groupID string, beforeCtime int64, beforeMessageID string, limit int) ([]ChatGroupMessageItem, error)
	ListMessagesAfter(ctx context.Context, groupID string, afterCtime int64, afterMessageID string, limit int) ([]ChatGroupMessageItem, error)
	ListDirectMessagesLatest(ctx context.Context, userID, partnerID string, limit int) ([]ChatDirectMessageItem, error)
	ListDirectMessagesBefore(ctx context.Context, userID, partnerID string, beforeCtime int64, beforeMessageID string, limit int) ([]ChatDirectMessageItem, error)
	ListDirectMessagesAfter(ctx context.Context, userID, partnerID string, afterCtime int64, afterMessageID string, limit int) ([]ChatDirectMessageItem, error)
	MarkDirectSeen(ctx context.Context, userID, partnerID string, lastSeenTime int64, now int64) error

	GetMemberRole(ctx context.Context, groupID, userID string) (string, error)
}

type ChatRepo struct {
	db  *pgxpool.Pool
	rdb redis.Cmdable
	log *zap.SugaredLogger
	dao *dao.Queries
}

func NewChatRepo(db *pgxpool.Pool, rdb redis.Cmdable, log *zap.SugaredLogger) *ChatRepo {
	return &ChatRepo{
		db:  db,
		rdb: rdb,
		log: zapx.OrNop(log),
		dao: dao.New(db),
	}
}

var _ IChatRepo = (*ChatRepo)(nil)

func (r *ChatRepo) commit(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Commit(ctx); err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return nil
}

func toString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return fmt.Sprint(t)
	}
}
