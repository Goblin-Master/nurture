package logic

import (
	"context"
	"nurture/internal/pkg/zapx"
	"nurture/internal/post/dto"
	"nurture/internal/post/repo"

	"go.uber.org/zap"
)

type FollowReader interface {
	IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error)
}

type IPostLogic interface {
	Home(ctx context.Context, userID string, req dto.PostHomeListReq) (dto.PostListResp, error)
	ListByTag(ctx context.Context, userID string, req dto.PostTagListReq) (dto.PostListResp, error)
	Search(ctx context.Context, userID string, req dto.PostSearchListReq) (dto.PostListResp, error)
	ListMyPosts(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error)
	ListMyDrafts(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error)
	ListMyMilestones(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error)
	Following(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error)
	Detail(ctx context.Context, userID string, req dto.PostDetailReq) (dto.PostDetailResp, error)
	NewPost(ctx context.Context, userID string, req dto.CreatePostReq) (dto.CreatePostResp, error)
	Publish(ctx context.Context, userID string, req dto.PublishPostReq) (dto.PublishPostResp, error)
	UpdateDraft(ctx context.Context, userID string, uri dto.PostDetailReq, req dto.UpdateDraftReq) (dto.UpdateDraftResp, error)
	DeleteDraft(ctx context.Context, userID string, uri dto.PostDetailReq) error
	DeletePost(ctx context.Context, userID string, uri dto.PostDetailReq) error
	CreateComment(ctx context.Context, userID string, postID string, req dto.CreateCommentReq) (dto.CreateCommentResp, error)
	LikePost(ctx context.Context, userID string, uri dto.PostDetailReq) error
	UnlikePost(ctx context.Context, userID string, uri dto.PostDetailReq) error
	ListComments(ctx context.Context, userID string, postID string, req dto.CommentListReq) (dto.CommentListResp, error)
	ListReplies(ctx context.Context, userID string, uri dto.CommentRepliesReq, req dto.CommentListReq) (dto.CommentListResp, error)
	DeleteComment(ctx context.Context, userID string, uri dto.CommentDeleteReq) error
	UpdateComment(ctx context.Context, userID string, uri dto.CommentUpdateReq, req dto.UpdateCommentReq) error
	LikeComment(ctx context.Context, userID string, uri dto.CommentLikeReq) error
	UnlikeComment(ctx context.Context, userID string, uri dto.CommentLikeReq) error
	CollectPost(ctx context.Context, userID string, uri dto.PostDetailReq) (dto.CollectResp, error)
	UncollectPost(ctx context.Context, userID string, uri dto.PostDetailReq) (dto.CollectResp, error)
	ListMyCollections(ctx context.Context, userID string, req dto.PostMyListReq) (dto.PostListResp, error)
	// admin tag
	AdminCreateTag(ctx context.Context, req dto.AdminTagCreateReq) (dto.AdminTagCreateResp, error)
	AdminDeleteTag(ctx context.Context, uri dto.AdminTagDeleteUri) error
	// public
	ListTags(ctx context.Context, req dto.TagListReq) (dto.TagListResp, error)
}

type PostLogic struct {
	postRepo     repo.IPostRepo
	followReader FollowReader
	log          *zap.SugaredLogger
}

func NewPostLogic(postRepo repo.IPostRepo, followReader FollowReader, log *zap.SugaredLogger) *PostLogic {
	return &PostLogic{
		postRepo:     postRepo,
		followReader: followReader,
		log:          zapx.OrNop(log),
	}
}

var _ IPostLogic = (*PostLogic)(nil)
