CREATE EXTENSION IF NOT EXISTS vector;

-- 1. 帖子表（PostgreSQL）
CREATE TABLE IF NOT EXISTS "post" (
  id               BIGSERIAL PRIMARY KEY,
  post_id          BIGINT UNIQUE NOT NULL,
  author_id        UUID NOT NULL,
  title            VARCHAR(255) NOT NULL,
  content          TEXT NOT NULL,
  status           VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published','milestone')),
  like_count       INTEGER NOT NULL DEFAULT 0,
  dislike_count    INTEGER NOT NULL DEFAULT 0,
  collect_count    INTEGER NOT NULL DEFAULT 0,
  comment_count    INTEGER NOT NULL DEFAULT 0,
  cover            VARCHAR(255) NOT NULL DEFAULT '',
  ctime            BIGINT NOT NULL,
  utime            BIGINT NOT NULL
);

COMMENT ON TABLE "post" IS '帖子表';
COMMENT ON COLUMN "post".id IS '主键ID';
COMMENT ON COLUMN "post".post_id IS '帖子ID';
COMMENT ON COLUMN "post".author_id IS '作者 user_id(UUID)';
COMMENT ON COLUMN "post".title IS '标题';
COMMENT ON COLUMN "post".content IS '正文';
COMMENT ON COLUMN "post".status IS '状态';
COMMENT ON COLUMN "post".like_count IS '点赞数';
COMMENT ON COLUMN "post".dislike_count IS '不喜欢数';
COMMENT ON COLUMN "post".collect_count IS '收藏数';
COMMENT ON COLUMN "post".comment_count IS '评论数';
COMMENT ON COLUMN "post".cover IS '封面图URL';
COMMENT ON COLUMN "post".ctime IS '创建时间戳';
COMMENT ON COLUMN "post".utime IS '更新时间戳';

-- 2. 标签表
CREATE TABLE IF NOT EXISTS "tag" (
  id           BIGSERIAL PRIMARY KEY,
  tag_id       BIGINT UNIQUE NOT NULL,
  tag_name     VARCHAR(64) UNIQUE NOT NULL,
  description  TEXT,
  ctime        BIGINT NOT NULL,
  utime        BIGINT NOT NULL
);

COMMENT ON TABLE "tag" IS '标签表';
COMMENT ON COLUMN "tag".id IS '主键ID';
COMMENT ON COLUMN "tag".tag_id IS '标签ID';
COMMENT ON COLUMN "tag".tag_name IS '标签名';
COMMENT ON COLUMN "tag".description IS '描述';
COMMENT ON COLUMN "tag".ctime IS '创建时间戳';
COMMENT ON COLUMN "tag".utime IS '更新时间戳';

-- 3. 帖子-标签关联表
CREATE TABLE IF NOT EXISTS "post_tag" (
  id        BIGSERIAL PRIMARY KEY,
  post_id   BIGINT NOT NULL,
  tag_id    BIGINT NOT NULL,
  CONSTRAINT uq_post_tag UNIQUE (post_id, tag_id)
);

COMMENT ON TABLE "post_tag" IS '帖子-标签关联表';
COMMENT ON COLUMN "post_tag".id IS '主键ID';
COMMENT ON COLUMN "post_tag".post_id IS '帖子ID';
COMMENT ON COLUMN "post_tag".tag_id IS '标签ID';

-- 4. 评论表（PostgreSQL）
CREATE TABLE IF NOT EXISTS "comment" (
  id          BIGSERIAL PRIMARY KEY,
  comment_id  BIGINT UNIQUE NOT NULL,
  post_id     BIGINT NOT NULL,
  user_id     UUID NOT NULL,
  parent_id   BIGINT,
  content     TEXT,
  ctime       BIGINT NOT NULL,
  utime       BIGINT NOT NULL
);

COMMENT ON TABLE "comment" IS '评论表';
COMMENT ON COLUMN "comment".id IS '主键ID';
COMMENT ON COLUMN "comment".comment_id IS '评论ID';
COMMENT ON COLUMN "comment".post_id IS '帖子ID';
COMMENT ON COLUMN "comment".user_id IS '评论者user_id(UUID)';
COMMENT ON COLUMN "comment".parent_id IS '父评论ID(一级为NULL)';
COMMENT ON COLUMN "comment".content IS '评论正文';
COMMENT ON COLUMN "comment".ctime IS '创建时间戳';
COMMENT ON COLUMN "comment".utime IS '更新时间戳';

-- 5. 评论闭包表
CREATE TABLE IF NOT EXISTS "comment_closure" (
  id          BIGSERIAL PRIMARY KEY,
  ancestor    BIGINT NOT NULL,
  descendant  BIGINT NOT NULL,
  depth       INTEGER NOT NULL CHECK (depth >= 0),
  CONSTRAINT uq_comment_closure UNIQUE (ancestor, descendant)
);

COMMENT ON TABLE "comment_closure" IS '评论闭包表';
COMMENT ON COLUMN "comment_closure".id IS '主键ID';
COMMENT ON COLUMN "comment_closure".ancestor IS '祖先评论ID';
COMMENT ON COLUMN "comment_closure".descendant IS '后代评论ID';
COMMENT ON COLUMN "comment_closure".depth IS '祖先到后代的距离';

-- 6. 点赞/不喜欢表
CREATE TABLE IF NOT EXISTS "like_dislike" (
  id       BIGSERIAL PRIMARY KEY,
  user_id  UUID NOT NULL,
  post_id  BIGINT NOT NULL,
  type     VARCHAR(20) NOT NULL CHECK (type IN ('like','dislike')),
  ctime    BIGINT NOT NULL,
  utime    BIGINT NOT NULL,
  CONSTRAINT uq_like_dislike UNIQUE (user_id, post_id)
);

COMMENT ON TABLE "like_dislike" IS '点赞/不喜欢表';
COMMENT ON COLUMN "like_dislike".id IS '主键ID';
COMMENT ON COLUMN "like_dislike".user_id IS '用户user_id(UUID)';
COMMENT ON COLUMN "like_dislike".post_id IS '帖子ID';
COMMENT ON COLUMN "like_dislike".type IS 'like或dislike';
COMMENT ON COLUMN "like_dislike".ctime IS '创建时间戳';
COMMENT ON COLUMN "like_dislike".utime IS '更新时间戳';

-- 7. 收藏表
CREATE TABLE IF NOT EXISTS "collection" (
  id             BIGSERIAL PRIMARY KEY,
  collection_id  BIGINT UNIQUE NOT NULL,
  user_id        UUID NOT NULL,
  post_id        BIGINT NOT NULL,
  ctime          BIGINT NOT NULL,
  utime          BIGINT NOT NULL,
  CONSTRAINT uq_collection UNIQUE (user_id, post_id)
);

COMMENT ON TABLE "collection" IS '收藏表';
COMMENT ON COLUMN "collection".id IS '主键ID';
COMMENT ON COLUMN "collection".collection_id IS '收藏ID';
COMMENT ON COLUMN "collection".user_id IS '用户user_id(UUID)';
COMMENT ON COLUMN "collection".post_id IS '帖子ID';
COMMENT ON COLUMN "collection".ctime IS '创建时间戳';
COMMENT ON COLUMN "collection".utime IS '更新时间戳';
