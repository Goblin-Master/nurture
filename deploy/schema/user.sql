-- 1. 扩展插件支持
CREATE EXTENSION IF NOT EXISTS vector;

-- 2. 用户表（PostgreSQL）
CREATE TABLE IF NOT EXISTS "user" (
  id        BIGSERIAL PRIMARY KEY,
  user_id   UUID UNIQUE NOT NULL, -- 直接用github.com/google/uuid生成的字符串
  ctime     BIGINT NOT NULL,
  utime     BIGINT NOT NULL,
  account   VARCHAR(20) UNIQUE NOT NULL,
  password  VARCHAR(20) NOT NULL,
  email     VARCHAR(20) UNIQUE NOT NULL,
  username  VARCHAR(20) NOT NULL,
  avatar    VARCHAR(255) NOT NULL,
  role      SMALLINT NOT NULL DEFAULT 1
);

COMMENT ON TABLE "user" IS '用户表';
COMMENT ON COLUMN "user".id IS '主键ID';
COMMENT ON COLUMN "user".user_id IS '用户ID';
COMMENT ON COLUMN "user".ctime IS '创建时间';
COMMENT ON COLUMN "user".utime IS '更新时间';
COMMENT ON COLUMN "user".account IS '账号';
COMMENT ON COLUMN "user".password IS '密码';
COMMENT ON COLUMN "user".email IS '邮箱';
COMMENT ON COLUMN "user".username IS '用户名';
COMMENT ON COLUMN "user".avatar IS '头像';
COMMENT ON COLUMN "user".role IS '角色';