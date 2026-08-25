# Post 模块主链路

本文档记录 `internal/post` 的主要业务链路。Post 模块拥有文章、标签、评论、点赞、收藏、关注 feed 和推荐画像的独立 SQL、repo/cache、handler/logic 边界。

## 模块启动链路

```mermaid
sequenceDiagram
  autonumber
  participant Router as router
  participant Module as post.Module
  participant Repo as post.repo
  participant Logic as post.logic
  participant Handler as post.handler
  participant AIX as pkg/aix
  participant User as user.Client via FollowReader

  Router->>Module: NewModule(DB, RDB, Log, AI, FollowReader)
  Module->>Repo: NewPostRepo(DB, RDB, Log, AI)
  Module->>Logic: NewPostLogic(repo, FollowReader, Log)
  Module->>Handler: NewPostHandler(logic)
  Router->>Module: RegisterRoutes(api.Group('/post'))
  Router->>Module: RegisterAdminRoutes(api.Group('/admin'))
  Repo-->>AIX: optional recommendation vector search/update
  Logic-->>User: read follow relationship when needed
```

## 草稿与发布链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as PostHandler
  participant Logic as PostLogic
  participant Repo as PostRepo
  participant DB as PostgreSQL
  participant Cache as post cache

  Client->>Handler: POST /api/post
  Handler->>Handler: auth user and bind CreatePostReq
  Handler->>Logic: NewPost(ctx, userID, req)
  Logic->>Logic: validate title, content and tags
  Logic->>Repo: CreatePost(postID, userID, draft content, tagIDs)
  Repo->>DB: begin transaction
  Repo->>DB: insert post draft
  Repo->>DB: insert post_tag rows
  Repo->>DB: commit
  Repo->>Cache: invalidate post list caches
  Logic-->>Handler: CreatePostResp

  Client->>Handler: POST /api/post/drafts/:post_id/publish
  Handler->>Logic: Publish(ctx, userID, postID)
  Logic->>Repo: Publish(postID, userID)
  Repo->>DB: update own draft status to published
  Repo->>Cache: invalidate post list/detail caches
  Logic-->>Handler: PublishPostResp
  Handler-->>Client: response
```

## 首页与搜索链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as PostHandler
  participant Logic as PostLogic
  participant Repo as PostRepo
  participant DB as PostgreSQL
  participant Cache as post cache
  participant AIX as pkg/aix

  Client->>Handler: GET /api/post?strategy=...
  Handler->>Logic: Home(ctx, userID, req)
  Logic->>Logic: normalize page and page_size
  Logic->>Repo: ListHome(userID, page, pageSize, strategy)
  alt recommendation strategy and AI available
    Repo->>AIX: SimilaritySearch(user recommend profile)
    AIX-->>Repo: recommended post IDs
    Repo->>DB: query recommended posts by IDs
  else ctime, hot, or random strategy
    Repo->>Cache: read list cache
    alt cache miss
      Repo->>DB: query published posts
      Repo->>Cache: cache list page
    end
  end
  Repo-->>Logic: post rows and hasMore
  Logic->>Logic: enrich response with baby age and viewer state
  Logic-->>Handler: PostListResp
  Handler-->>Client: response

  Client->>Handler: GET /api/post/search
  Handler->>Logic: Search(ctx, userID, req)
  Logic->>Repo: Search(userID, keyword, tagID, strategy, page, pageSize)
  Repo->>DB: query title/tag filtered posts
  Repo-->>Logic: post rows and hasMore
  Logic-->>Handler: PostListResp
```

## 互动与推荐画像链路

```mermaid
sequenceDiagram
  autonumber
  participant Client as client
  participant Handler as PostHandler
  participant Logic as PostLogic
  participant Repo as PostRepo
  participant DB as PostgreSQL
  participant Cache as post cache
  participant AIX as pkg/aix

  Client->>Handler: POST /api/post/:post_id/like
  Handler->>Logic: LikePost(ctx, userID, postID)
  Logic->>Repo: LikePost(postID, userID)
  Repo->>DB: insert like row and update counts
  Repo->>Cache: invalidate post detail/list caches
  Logic->>Repo: TouchUserRecommendProfile(userID, postID)
  Repo->>AIX: add or update recommendation profile document when enabled
  Logic->>Repo: TouchUserTagPref(userID, postID, +3)
  Repo->>DB: upsert user tag preference
  Logic-->>Handler: nil

  Client->>Handler: POST /api/post/:post_id/comments
  Handler->>Logic: CreateComment(ctx, userID, postID, req)
  Logic->>Repo: CreateComment(commentID, postID, userID, parentID, content)
  Repo->>DB: begin transaction
  Repo->>DB: insert comment and closure rows
  Repo->>DB: update post comment count
  Repo->>DB: commit
  Repo->>Cache: invalidate comment and post caches
  Logic->>Repo: TouchUserRecommendProfile(userID, postID)
  Logic-->>Handler: CreateCommentResp
```

## 边界说明

- Post 模块通过注入的 `FollowReader` 读取用户关注关系；当前由 `user.Client` 提供实现，不直接依赖 user 模块内部 repo/logic。
- 推荐依赖注入的 `pkg/aix`，AI 不可用时 feed/search 仍保持基础查询能力。
- 标签管理在模块内提供 admin routes，顶层 router 只负责挂到 `/api/admin`。
