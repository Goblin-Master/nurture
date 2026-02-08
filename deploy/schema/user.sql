-- 1. 扩展插件支持
CREATE EXTENSION IF NOT EXISTS vector;

-- 2. 用户基础信息（注册必填）
CREATE TABLE IF NOT EXISTS "user_base" (
  id        BIGSERIAL PRIMARY KEY,
  user_id   UUID UNIQUE NOT NULL, -- 直接用github.com/google/uuid生成的字符串
  account   VARCHAR(20) UNIQUE NOT NULL,
  password  VARCHAR(20) NOT NULL,
  email     VARCHAR(20) UNIQUE NOT NULL,
  username  VARCHAR(20) NOT NULL,
  gender    VARCHAR(10) NOT NULL CHECK (gender IN ('male','female')),
  role      SMALLINT NOT NULL DEFAULT 1,
  ctime     BIGINT NOT NULL,
  utime     BIGINT NOT NULL,
);

COMMENT ON TABLE "user_base" IS '用户基础信息（注册必填）';
COMMENT ON COLUMN "user_base".id IS '主键ID';
COMMENT ON COLUMN "user_base".user_id IS '用户ID';
COMMENT ON COLUMN "user_base".account IS '账号';
COMMENT ON COLUMN "user_base".password IS '密码';
COMMENT ON COLUMN "user_base".email IS '邮箱';
COMMENT ON COLUMN "user_base".username IS '用户名';
COMMENT ON COLUMN "user_base".gender IS '性别:male/female';
COMMENT ON COLUMN "user_base".role IS '角色';
COMMENT ON COLUMN "user_base".ctime IS '创建时间';
COMMENT ON COLUMN "user_base".utime IS '更新时间';

-- 3. 用户扩展信息（头像与地区，冗余用户名）
CREATE TABLE IF NOT EXISTS "user_addition" (
  id       BIGSERIAL PRIMARY KEY,
  user_id  UUID UNIQUE NOT NULL,
  username VARCHAR(20) NOT NULL,
  gender   VARCHAR(10) NOT NULL CHECK (gender IN ('male','female')),
  phone    VARCHAR(20) NOT NULL DEFAULT '',
  occupation VARCHAR(50) NOT NULL DEFAULT '',
  birthday BIGINT NOT NULL DEFAULT 0,
  avatar   VARCHAR(255) NOT NULL DEFAULT '',
  province VARCHAR(50)  NOT NULL DEFAULT '',
  city     VARCHAR(50)  NOT NULL DEFAULT '',
  ctime    BIGINT NOT NULL,
  utime    BIGINT NOT NULL
);

COMMENT ON TABLE "user_addition" IS '用户扩展信息（冗余用户名、头像、地区）';
COMMENT ON COLUMN "user_addition".id IS '主键ID';
COMMENT ON COLUMN "user_addition".user_id IS '用户ID(与基础信息一对一)';
COMMENT ON COLUMN "user_addition".username IS '用户名（冗余）';
COMMENT ON COLUMN "user_addition".gender IS '性别（冗余）';
COMMENT ON COLUMN "user_addition".phone IS '手机号';
COMMENT ON COLUMN "user_addition".occupation IS '职业';
COMMENT ON COLUMN "user_addition".birthday IS '生日(毫秒时间戳,未知为0)';
COMMENT ON COLUMN "user_addition".avatar IS '头像URL';
COMMENT ON COLUMN "user_addition".province IS '省';
COMMENT ON COLUMN "user_addition".city IS '市';
COMMENT ON COLUMN "user_addition".ctime IS '创建时间';
COMMENT ON COLUMN "user_addition".utime IS '更新时间';

-- 4. 用户另一半关系表
CREATE TABLE IF NOT EXISTS "user_partner" (
  id        BIGSERIAL PRIMARY KEY,
  father    UUID NOT NULL,
  mother    UUID NOT NULL,
  ctime     BIGINT NOT NULL,
  utime     BIGINT NOT NULL,
  CONSTRAINT ck_partner_users CHECK (father <> mother),
  CONSTRAINT uq_partner UNIQUE (father, mother)
);

COMMENT ON TABLE "user_partner" IS '用户另一半关系表（父/母）';
COMMENT ON COLUMN "user_partner".id IS '主键ID';
COMMENT ON COLUMN "user_partner".father IS '父亲用户UUID';
COMMENT ON COLUMN "user_partner".mother IS '母亲用户UUID';
COMMENT ON COLUMN "user_partner".ctime IS '创建时间戳';
COMMENT ON COLUMN "user_partner".utime IS '更新时间戳';

-- 5. 插入默认管理员账号（同时写入基础与扩展）
WITH u AS (
  INSERT INTO "user_base" (user_id, ctime, utime, account, password, email, username, gender, role)
  VALUES (
    gen_random_uuid(),
    EXTRACT(EPOCH FROM NOW()) * 1000,
    EXTRACT(EPOCH FROM NOW()) * 1000,
    'admin',
    'admin',
    'admin@nurture.com',
    'Admin',
    'male',
    3
  )
  ON CONFLICT (account) DO NOTHING
  RETURNING user_id, username, gender, ctime, utime
)
INSERT INTO "user_addition" (user_id, username, gender, phone, occupation, birthday, avatar, province, city, ctime, utime)
SELECT user_id, username, gender, '', '', 0, '', '', '', ctime, utime
FROM u;
