CREATE TABLE IF NOT EXISTS "chat_group" (
  id           BIGSERIAL PRIMARY KEY,
  group_id     UUID UNIQUE NOT NULL,
  owner_id     UUID NOT NULL,
  name         VARCHAR(100) NOT NULL,
  avatar       VARCHAR(255) NOT NULL DEFAULT '',
  description  TEXT NOT NULL DEFAULT '',
  member_limit INTEGER NOT NULL,
  member_count INTEGER NOT NULL DEFAULT 0,
  ctime        BIGINT NOT NULL,
  utime        BIGINT NOT NULL,
  CONSTRAINT ck_chat_group_member_limit CHECK (member_limit > 0),
  CONSTRAINT ck_chat_group_member_count CHECK (member_count >= 0 AND member_count <= member_limit)
);

COMMENT ON TABLE "chat_group" IS '群聊-群表';
COMMENT ON COLUMN "chat_group".id IS '主键ID';
COMMENT ON COLUMN "chat_group".group_id IS '群ID';
COMMENT ON COLUMN "chat_group".owner_id IS '群主用户ID(UUID)';
COMMENT ON COLUMN "chat_group".name IS '群名称';
COMMENT ON COLUMN "chat_group".avatar IS '群头像URL';
COMMENT ON COLUMN "chat_group".description IS '群描述';
COMMENT ON COLUMN "chat_group".member_limit IS '成员人数上限';
COMMENT ON COLUMN "chat_group".member_count IS '成员人数(冗余计数)';
COMMENT ON COLUMN "chat_group".ctime IS '创建时间戳(毫秒)';
COMMENT ON COLUMN "chat_group".utime IS '更新时间戳(毫秒)';

CREATE TABLE IF NOT EXISTS "chat_group_member" (
  id        BIGSERIAL PRIMARY KEY,
  group_id  UUID NOT NULL,
  user_id   UUID NOT NULL,
  role      VARCHAR(10) NOT NULL CHECK (role IN ('owner','admin','member')),
  last_seen_time BIGINT NOT NULL DEFAULT 0,
  ctime     BIGINT NOT NULL,
  utime     BIGINT NOT NULL,
  CONSTRAINT uq_chat_group_member UNIQUE (group_id, user_id)
);

COMMENT ON TABLE "chat_group_member" IS '群聊-群成员表';
COMMENT ON COLUMN "chat_group_member".id IS '主键ID';
COMMENT ON COLUMN "chat_group_member".group_id IS '群ID';
COMMENT ON COLUMN "chat_group_member".user_id IS '用户ID(UUID)';
COMMENT ON COLUMN "chat_group_member".role IS '成员角色(owner/admin/member)';
COMMENT ON COLUMN "chat_group_member".last_seen_time IS '最后一次离开群详情页时间(毫秒,用于未读统计)';
COMMENT ON COLUMN "chat_group_member".ctime IS '创建时间戳(毫秒)';
COMMENT ON COLUMN "chat_group_member".utime IS '更新时间戳(毫秒)';

CREATE INDEX IF NOT EXISTS idx_chat_group_member_user ON "chat_group_member"(user_id, group_id);
CREATE INDEX IF NOT EXISTS idx_chat_group_member_group ON "chat_group_member"(group_id, user_id);

CREATE TABLE IF NOT EXISTS "chat_group_message" (
  id           BIGSERIAL PRIMARY KEY,
  message_id   UUID NOT NULL,
  group_id     UUID NOT NULL,
  from_user_id UUID NOT NULL,
  type         VARCHAR(20) NOT NULL CHECK (type IN ('text','image','system')),
  content      TEXT NOT NULL,
  ctime        BIGINT NOT NULL,
  utime        BIGINT NOT NULL,
  CONSTRAINT uq_chat_group_message UNIQUE (group_id, message_id)
);

COMMENT ON TABLE "chat_group_message" IS '群聊-群消息表';
COMMENT ON COLUMN "chat_group_message".id IS '主键ID';
COMMENT ON COLUMN "chat_group_message".message_id IS '消息ID(UUID,幂等)';
COMMENT ON COLUMN "chat_group_message".group_id IS '群ID';
COMMENT ON COLUMN "chat_group_message".from_user_id IS '发送者用户ID(UUID)';
COMMENT ON COLUMN "chat_group_message".type IS '消息类型(text/image/system)';
COMMENT ON COLUMN "chat_group_message".content IS '消息内容(文本或图片URL)';
COMMENT ON COLUMN "chat_group_message".ctime IS '创建时间戳(毫秒)';
COMMENT ON COLUMN "chat_group_message".utime IS '更新时间戳(毫秒)';

CREATE INDEX IF NOT EXISTS idx_chat_group_message_group_ctime ON "chat_group_message"(group_id, ctime, message_id);
