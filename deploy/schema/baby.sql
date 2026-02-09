CREATE EXTENSION IF NOT EXISTS vector;

-- 1. 宝宝表 (PostgreSQL)
CREATE TABLE IF NOT EXISTS "baby" (
  id          BIGSERIAL PRIMARY KEY,
  baby_id     UUID NOT NULL,
  user_id     UUID NOT NULL,
  name        VARCHAR(50) NOT NULL,
  gender      VARCHAR(10) NOT NULL CHECK (gender IN ('male','female')),
  birthday    BIGINT NOT NULL,
  avatar      VARCHAR(255) NOT NULL DEFAULT '',
  ctime       BIGINT NOT NULL,
  utime       BIGINT NOT NULL,
  CONSTRAINT uq_baby_user UNIQUE (baby_id, user_id)
);

COMMENT ON TABLE "baby" IS '宝宝表';
COMMENT ON COLUMN "baby".id IS '主键ID';
COMMENT ON COLUMN "baby".baby_id IS '宝宝ID';
COMMENT ON COLUMN "baby".user_id IS '所属用户ID(UUID)';
COMMENT ON COLUMN "baby".name IS '宝宝姓名';
COMMENT ON COLUMN "baby".gender IS '性别:male/female';
COMMENT ON COLUMN "baby".birthday IS '生日(毫秒时间戳)';
COMMENT ON COLUMN "baby".avatar IS '宝宝头像URL';
COMMENT ON COLUMN "baby".ctime IS '创建时间戳';
COMMENT ON COLUMN "baby".utime IS '更新时间戳';

-- 2. 疫苗字典表
CREATE TABLE IF NOT EXISTS "vaccine" (
  id                   BIGSERIAL PRIMARY KEY,
  vaccine_id           UUID UNIQUE NOT NULL,
  name                 VARCHAR(100) UNIQUE NOT NULL,
  disease              VARCHAR(255) NOT NULL,
  recommend_age_days   INTEGER NOT NULL,
  link                 VARCHAR(255) NOT NULL DEFAULT '',
  ctime                BIGINT NOT NULL,
  utime                BIGINT NOT NULL
);

COMMENT ON TABLE "vaccine" IS '疫苗字典表';
COMMENT ON COLUMN "vaccine".id IS '主键ID';
COMMENT ON COLUMN "vaccine".vaccine_id IS '疫苗ID';
COMMENT ON COLUMN "vaccine".name IS '疫苗名称';
COMMENT ON COLUMN "vaccine".disease IS '预防疾病';
COMMENT ON COLUMN "vaccine".recommend_age_days IS '推荐接种年龄(天)';
COMMENT ON COLUMN "vaccine".link IS '详情页面URL';
COMMENT ON COLUMN "vaccine".ctime IS '创建时间戳';
COMMENT ON COLUMN "vaccine".utime IS '更新时间戳';

-- 3. 宝宝疫苗接种记录表 (PostgreSQL)
CREATE TABLE IF NOT EXISTS "baby_vaccine_record" (
  id                   BIGSERIAL PRIMARY KEY,
  record_id            UUID UNIQUE NOT NULL,
  baby_id              UUID NOT NULL,
  vaccine_id           UUID NOT NULL,
  due_time             BIGINT NOT NULL,
  status               VARCHAR(20) NOT NULL CHECK (status IN ('given','not_given')),
  actual_time          BIGINT,
  ctime                BIGINT NOT NULL,
  utime                BIGINT NOT NULL,
  CONSTRAINT uq_bvr UNIQUE (baby_id, vaccine_id),
  CONSTRAINT ck_bvr_status_time CHECK (
    (status = 'given' AND actual_time IS NOT NULL) OR
    (status = 'not_given' AND actual_time IS NULL)
  )
);

COMMENT ON TABLE "baby_vaccine_record" IS '宝宝疫苗接种记录表';
COMMENT ON COLUMN "baby_vaccine_record".id IS '主键ID';
COMMENT ON COLUMN "baby_vaccine_record".record_id IS '记录ID';
COMMENT ON COLUMN "baby_vaccine_record".baby_id IS '宝宝ID(外键)';
COMMENT ON COLUMN "baby_vaccine_record".vaccine_id IS '疫苗ID(外键)';
COMMENT ON COLUMN "baby_vaccine_record".due_time IS '应接种时间(毫秒)';
COMMENT ON COLUMN "baby_vaccine_record".status IS '接种状态';
COMMENT ON COLUMN "baby_vaccine_record".actual_time IS '实际接种时间(毫秒)';
COMMENT ON COLUMN "baby_vaccine_record".ctime IS '创建时间戳';
COMMENT ON COLUMN "baby_vaccine_record".utime IS '更新时间戳';
CREATE INDEX IF NOT EXISTS idx_bvr_baby ON "baby_vaccine_record"(baby_id);
CREATE INDEX IF NOT EXISTS idx_bvr_baby_status ON "baby_vaccine_record"(baby_id, status);

-- 4. 宝宝成长记录表 (PostgreSQL)
CREATE TABLE IF NOT EXISTS "baby_growth_record" (
  id                  BIGSERIAL PRIMARY KEY,
  record_id           UUID UNIQUE NOT NULL,
  baby_id             UUID NOT NULL,
  user_id             UUID NOT NULL,
  record_time         BIGINT NOT NULL,
  height              DOUBLE PRECISION,
  weight              DOUBLE PRECISION,
  head_circumference  DOUBLE PRECISION,
  remark              TEXT,
  ctime               BIGINT NOT NULL,
  utime               BIGINT NOT NULL
);

COMMENT ON TABLE "baby_growth_record" IS '宝宝成长记录表';
COMMENT ON COLUMN "baby_growth_record".id IS '主键ID';
COMMENT ON COLUMN "baby_growth_record".record_id IS '记录ID';
COMMENT ON COLUMN "baby_growth_record".baby_id IS '宝宝ID(外键)';
COMMENT ON COLUMN "baby_growth_record".user_id IS '用户ID(UUID)';
COMMENT ON COLUMN "baby_growth_record".record_time IS '记录时间(毫秒时间戳)';
COMMENT ON COLUMN "baby_growth_record".height IS '身高(cm)';
COMMENT ON COLUMN "baby_growth_record".weight IS '体重(kg)';
COMMENT ON COLUMN "baby_growth_record".head_circumference IS '头围(cm)';
COMMENT ON COLUMN "baby_growth_record".remark IS '备注';
COMMENT ON COLUMN "baby_growth_record".ctime IS '创建时间戳';
COMMENT ON COLUMN "baby_growth_record".utime IS '更新时间戳';
CREATE INDEX IF NOT EXISTS idx_bgr_baby_time ON "baby_growth_record"(baby_id, record_time);

-- 5. 宝宝相册表 (PostgreSQL)
CREATE TABLE IF NOT EXISTS "baby_photo" (
  id          BIGSERIAL PRIMARY KEY,
  photo_id    UUID UNIQUE NOT NULL,
  baby_id     UUID NOT NULL,
  link        VARCHAR(255) NOT NULL DEFAULT '',
  caption     TEXT,
  ctime       BIGINT NOT NULL,
  utime       BIGINT NOT NULL
);

COMMENT ON TABLE "baby_photo" IS '宝宝相册表';
COMMENT ON COLUMN "baby_photo".id IS '主键ID';
COMMENT ON COLUMN "baby_photo".photo_id IS '相片ID';
COMMENT ON COLUMN "baby_photo".baby_id IS '宝宝ID(外键)';
COMMENT ON COLUMN "baby_photo".link IS '详情页面URL';
COMMENT ON COLUMN "baby_photo".caption IS '描述';
COMMENT ON COLUMN "baby_photo".ctime IS '创建时间戳';
COMMENT ON COLUMN "baby_photo".utime IS '更新时间戳';
