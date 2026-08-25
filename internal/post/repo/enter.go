package repo

import (
	"context"
	"nurture/internal/pkg/aix"
	"nurture/internal/pkg/zapx"
	"nurture/internal/post/repo/dao"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PostRow struct {
	PostID         string
	AuthorID       string
	AuthorName     string
	AuthorAvatar   string
	AuthorProvince string
	AuthorCity     string
	Title          string
	Content        string
	Status         string
	LikeCount      int32
	DislikeCount   int32
	CollectCount   int32
	CommentCount   int32
	Ctime          int64
	Utime          int64
	Birthday       int64
	Tags           []string
	IsLike         bool
	IsDislike      bool
	IsCollect      bool
}

type TagRow struct {
	TagID       string
	Name        string
	Description string
}

type IPostRepo interface {
	ListHome(ctx context.Context, userID string, page, pageSize int, strategy string) ([]PostRow, bool, error)
	ListByTag(ctx context.Context, userID, tagID string, page, pageSize int, strategy string) ([]PostRow, bool, error)
	Search(ctx context.Context, userID, keyword, tagID, strategy string, page, pageSize int) ([]PostRow, bool, error)
	ListByAuthor(ctx context.Context, authorID string, page, pageSize int, strategy string) ([]PostRow, bool, error)
	ListDraftsByAuthor(ctx context.Context, authorID string, page, pageSize int) ([]PostRow, bool, error)
	ListMilestonesByAuthor(ctx context.Context, authorID string, page, pageSize int) ([]PostRow, bool, error)
	ListFollowing(ctx context.Context, userID string, page, pageSize int, strategy string) ([]PostRow, bool, error)
	GetDetail(ctx context.Context, userID, postID string) (PostRow, error)
	CreatePost(ctx context.Context, postID, authorID, title, content, status string, ctime, utime int64, tagIDs []string) error
	Publish(ctx context.Context, postID, userID string) error
	UpdateDraft(ctx context.Context, postID, userID, title, content string, tagIDs []string) error
	DeleteDraft(ctx context.Context, postID, authorID string) error
	DeletePost(ctx context.Context, postID, authorID string) error
	CreateComment(ctx context.Context, commentID, postID, userID string, parentID *string, content string, now int64) error
	GetPostStatus(ctx context.Context, postID string) (string, error)
	GetCommentParentInfo(ctx context.Context, commentID string) (string, string, error)
	ListCommentsByPost(ctx context.Context, postID string, userID string, page, pageSize int, strategy string) ([]CommentRow, bool, error)
	ListRepliesByComment(ctx context.Context, commentID string, userID string, page, pageSize int, strategy string) ([]CommentRow, bool, error)
	DeleteComment(ctx context.Context, commentID, userID string) error
	UpdateComment(ctx context.Context, commentID, userID, content string) error
	LikePost(ctx context.Context, postID, userID string) error
	UnlikePost(ctx context.Context, postID, userID string) error
	LikeComment(ctx context.Context, commentID, userID string) error
	UnlikeComment(ctx context.Context, commentID, userID string) error
	CollectPost(ctx context.Context, postID, userID, collectionID string) error
	UncollectPost(ctx context.Context, postID, userID string) error
	ListMyCollections(ctx context.Context, userID string, page, pageSize int, strategy string) ([]PostRow, bool, error)
	TouchUserRecommendProfile(ctx context.Context, userID string, postID string) error
	IndexPostForRecommend(ctx context.Context, postID string) error
	TouchUserTagPref(ctx context.Context, userID string, postID string, score float64) error
	// admin tag
	CreateTag(ctx context.Context, tagID, name, description string, now int64) (TagRow, error)
	DeleteTag(ctx context.Context, tagID string) error
	ListTags(ctx context.Context, keyword string, page, pageSize int) ([]TagRow, bool, error)
}

type PostRepo struct {
	db  *pgxpool.Pool
	dao *dao.Queries
	rdb redis.Cmdable
	log *zap.SugaredLogger
	ai  *aix.AIX
}

func NewPostRepo(db *pgxpool.Pool, rdb redis.Cmdable, log *zap.SugaredLogger, ai *aix.AIX) *PostRepo {
	return &PostRepo{
		db:  db,
		dao: dao.New(db),
		rdb: rdb,
		log: zapx.OrNop(log),
		ai:  ai,
	}
}

var _ IPostRepo = (*PostRepo)(nil)
