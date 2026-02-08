CREATE EXTENSION IF NOT EXISTS vector;

-- 1. 喂养记录表 (PostgreSQL)
CREATE TABLE IF NOT EXISTS "daily_feeding" (
  id            BIGSERIAL PRIMARY KEY,
  feeding_id    BIGINT UNIQUE NOT NULL,
  baby_id       BIGINT NOT NULL,
  user_id       UUID NOT NULL,
  feed_time     BIGINT NOT NULL,
  feed_type     VARCHAR(20) NOT NULL CHECK (feed_type IN ('breast_milk','formula','solid')),
  remark        TEXT,
  ctime         BIGINT NOT NULL,
  utime         BIGINT NOT NULL
);

COMMENT ON TABLE "daily_feeding" IS '喂养记录表';
COMMENT ON COLUMN "daily_feeding".id IS '主键ID';
COMMENT ON COLUMN "daily_feeding".feeding_id IS '喂养记录ID';
COMMENT ON COLUMN "daily_feeding".baby_id IS '宝宝ID(外键)';
COMMENT ON COLUMN "daily_feeding".user_id IS '用户ID(UUID)';
COMMENT ON COLUMN "daily_feeding".feed_time IS '喂养时间(毫秒时间戳)';
COMMENT ON COLUMN "daily_feeding".feed_type IS '喂养方式';
COMMENT ON COLUMN "daily_feeding".remark IS '备注';
COMMENT ON COLUMN "daily_feeding".ctime IS '创建时间戳';
COMMENT ON COLUMN "daily_feeding".utime IS '更新时间戳';

-- 2. 睡眠记录表 (PostgreSQL)
CREATE TABLE IF NOT EXISTS "daily_sleep" (
  id           BIGSERIAL PRIMARY KEY,
  sleep_id     BIGINT UNIQUE NOT NULL,
  baby_id      BIGINT NOT NULL,
  user_id      UUID NOT NULL,
  start_time   BIGINT NOT NULL,
  end_time     BIGINT,
  duration     BIGINT,
  source       VARCHAR(10) NOT NULL CHECK (source IN ('timer','manual')),
  status       VARCHAR(10) NOT NULL CHECK (status IN ('running','finished')),
  remark       TEXT,
  ctime        BIGINT NOT NULL,
  utime        BIGINT NOT NULL,
 
  CONSTRAINT ck_ds_time CHECK (
    (status = 'running'  AND end_time IS NULL)
    OR
    (status = 'finished' AND end_time IS NOT NULL AND end_time > start_time)
  )
);

COMMENT ON TABLE "daily_sleep" IS '睡眠记录表';
COMMENT ON COLUMN "daily_sleep".id IS '主键ID';
COMMENT ON COLUMN "daily_sleep".sleep_id IS '睡眠记录ID';
COMMENT ON COLUMN "daily_sleep".baby_id IS '宝宝ID(外键)';
COMMENT ON COLUMN "daily_sleep".user_id IS '用户ID(UUID)';
COMMENT ON COLUMN "daily_sleep".start_time IS '开始时间(毫秒时间戳)';
COMMENT ON COLUMN "daily_sleep".end_time IS '结束时间(毫秒时间戳)';
COMMENT ON COLUMN "daily_sleep".duration IS '睡眠时长(毫秒)';
COMMENT ON COLUMN "daily_sleep".source IS '记录方式(timer/manual)';
COMMENT ON COLUMN "daily_sleep".status IS '状态(running/finished)';
COMMENT ON COLUMN "daily_sleep".remark IS '备注';
COMMENT ON COLUMN "daily_sleep".ctime IS '创建时间戳';
COMMENT ON COLUMN "daily_sleep".utime IS '更新时间戳';

-- 3. 换尿布记录表 (PostgreSQL)
CREATE TABLE IF NOT EXISTS "daily_diaper" (
  id               BIGSERIAL PRIMARY KEY,
  diaper_id        BIGINT UNIQUE NOT NULL,
  baby_id          BIGINT NOT NULL,
  user_id          UUID NOT NULL,
  change_time      BIGINT NOT NULL,
  diaper_type      VARCHAR(10) NOT NULL CHECK (diaper_type IN ('pee','poop','both','dry')),
  pee_color        VARCHAR(20) CHECK (pee_color IN ('milky_white','pink','normal','yellow','red','tea')),
  poop_color       VARCHAR(20) CHECK (poop_color IN ('dark_green','green','yellow','orange','red','black','gray_white')),
  poop_consistency VARCHAR(20) CHECK (poop_consistency IN ('paste','foamy','milky','food_residue','egg_flower','watery','sheep','bloody')),
  remark           TEXT,
  ctime            BIGINT NOT NULL,
  utime            BIGINT NOT NULL,
  CONSTRAINT ck_dd_fields CHECK (
    (diaper_type = 'pee'  AND pee_color IS NOT NULL AND poop_color IS NULL AND poop_consistency IS NULL)
    OR
    (diaper_type = 'poop' AND pee_color IS NULL AND poop_color IS NOT NULL AND poop_consistency IS NOT NULL)
    OR
    (diaper_type = 'both' AND pee_color IS NOT NULL AND poop_color IS NOT NULL AND poop_consistency IS NOT NULL)
    OR
    (diaper_type = 'dry'  AND pee_color IS NULL AND poop_color IS NULL AND poop_consistency IS NULL)
  )
);

COMMENT ON TABLE "daily_diaper" IS '换尿布记录表';
COMMENT ON COLUMN "daily_diaper".id IS '主键ID';
COMMENT ON COLUMN "daily_diaper".diaper_id IS '换尿布记录ID';
COMMENT ON COLUMN "daily_diaper".baby_id IS '宝宝ID(外键)';
COMMENT ON COLUMN "daily_diaper".user_id IS '用户ID(UUID)';
COMMENT ON COLUMN "daily_diaper".change_time IS '更换时间(毫秒时间戳)';
COMMENT ON COLUMN "daily_diaper".diaper_type IS '尿布状态';
COMMENT ON COLUMN "daily_diaper".pee_color IS '嘘嘘颜色';
COMMENT ON COLUMN "daily_diaper".poop_color IS '便便颜色';
COMMENT ON COLUMN "daily_diaper".poop_consistency IS '便便状态';
COMMENT ON COLUMN "daily_diaper".remark IS '备注';
COMMENT ON COLUMN "daily_diaper".ctime IS '创建时间戳';
COMMENT ON COLUMN "daily_diaper".utime IS '更新时间戳';
