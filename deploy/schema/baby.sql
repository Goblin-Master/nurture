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
  link                 VARCHAR(255) NOT NULL DEFAULT '',
  ctime                BIGINT NOT NULL,
  utime                BIGINT NOT NULL
);

COMMENT ON TABLE "vaccine" IS '疫苗字典表';
COMMENT ON COLUMN "vaccine".id IS '主键ID';
COMMENT ON COLUMN "vaccine".vaccine_id IS '疫苗ID';
COMMENT ON COLUMN "vaccine".name IS '疫苗名称';
COMMENT ON COLUMN "vaccine".disease IS '预防疾病';
COMMENT ON COLUMN "vaccine".link IS '详情页面URL';
COMMENT ON COLUMN "vaccine".ctime IS '创建时间戳';
COMMENT ON COLUMN "vaccine".utime IS '更新时间戳';

-- 2.1 疫苗剂次表
CREATE TABLE IF NOT EXISTS "vaccine_dose" (
  id                   BIGSERIAL PRIMARY KEY,
  dose_id              UUID UNIQUE NOT NULL,
  vaccine_id           UUID NOT NULL,
  dose_number          INTEGER NOT NULL,
  recommend_age_days   INTEGER NOT NULL,
  ctime                BIGINT NOT NULL,
  utime                BIGINT NOT NULL,
  CONSTRAINT uq_vaccine_dose UNIQUE (vaccine_id, dose_number)
);

COMMENT ON TABLE "vaccine_dose" IS '疫苗剂次表';
COMMENT ON COLUMN "vaccine_dose".id IS '主键ID';
COMMENT ON COLUMN "vaccine_dose".dose_id IS '剂次ID';
COMMENT ON COLUMN "vaccine_dose".vaccine_id IS '疫苗ID(外键)';
COMMENT ON COLUMN "vaccine_dose".dose_number IS '第几剂(从1开始)';
COMMENT ON COLUMN "vaccine_dose".recommend_age_days IS '推荐接种年龄(天)';
COMMENT ON COLUMN "vaccine_dose".ctime IS '创建时间戳';
COMMENT ON COLUMN "vaccine_dose".utime IS '更新时间戳';
CREATE INDEX IF NOT EXISTS idx_vd_vaccine_number ON "vaccine_dose"(vaccine_id, dose_number);

-- 3. 宝宝疫苗接种记录表 (PostgreSQL)
CREATE TABLE IF NOT EXISTS "baby_vaccine_record" (
  id                   BIGSERIAL PRIMARY KEY,
  record_id            UUID UNIQUE NOT NULL,
  baby_id              UUID NOT NULL,
  dose_id              UUID NOT NULL,
  due_time             BIGINT NOT NULL,
  status               VARCHAR(20) NOT NULL CHECK (status IN ('given','not_given')),
  actual_time          BIGINT,
  ctime                BIGINT NOT NULL,
  utime                BIGINT NOT NULL,
  CONSTRAINT uq_bvr UNIQUE (baby_id, dose_id),
  CONSTRAINT ck_bvr_status_time CHECK (
    (status = 'given' AND actual_time IS NOT NULL) OR
    (status = 'not_given' AND actual_time IS NULL)
  )
);

COMMENT ON TABLE "baby_vaccine_record" IS '宝宝疫苗接种记录表';
COMMENT ON COLUMN "baby_vaccine_record".id IS '主键ID';
COMMENT ON COLUMN "baby_vaccine_record".record_id IS '记录ID';
COMMENT ON COLUMN "baby_vaccine_record".baby_id IS '宝宝ID(外键)';
COMMENT ON COLUMN "baby_vaccine_record".dose_id IS '剂次ID(外键)';
COMMENT ON COLUMN "baby_vaccine_record".due_time IS '应接种时间(毫秒)';
COMMENT ON COLUMN "baby_vaccine_record".status IS '接种状态';
COMMENT ON COLUMN "baby_vaccine_record".actual_time IS '实际接种时间(毫秒)';
COMMENT ON COLUMN "baby_vaccine_record".ctime IS '创建时间戳';
COMMENT ON COLUMN "baby_vaccine_record".utime IS '更新时间戳';
CREATE INDEX IF NOT EXISTS idx_bvr_baby ON "baby_vaccine_record"(baby_id);
CREATE INDEX IF NOT EXISTS idx_bvr_baby_status ON "baby_vaccine_record"(baby_id, status);
CREATE INDEX IF NOT EXISTS idx_bvr_baby_dose ON "baby_vaccine_record"(baby_id, dose_id);

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

-- 6. 初始化疫苗与剂次数据
INSERT INTO "vaccine" (vaccine_id, name, disease, link, ctime, utime)
VALUES (gen_random_uuid(), '乙肝疫苗', '预防乙型肝炎', 'https://baike.baidu.com/item/%E4%B9%99%E8%82%9D%E7%96%AB%E8%8B%97?fromModule=lemma_search-box', EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000)
ON CONFLICT (name) DO NOTHING;

INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 1, 0, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='乙肝疫苗';
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 2, 30, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='乙肝疫苗';
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 3, 180, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='乙肝疫苗';

INSERT INTO "vaccine" (vaccine_id, name, disease, link, ctime, utime)
VALUES (gen_random_uuid(), '卡介苗', '预防结核病', 'https://baike.baidu.com/item/%E5%8D%A1%E4%BB%8B%E8%8B%97?fromModule=lemma_search-box', EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000)
ON CONFLICT (name) DO NOTHING;
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 1, 0, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='卡介苗';

INSERT INTO "vaccine" (vaccine_id, name, disease, link, ctime, utime)
VALUES (gen_random_uuid(), '脊髓灰质炎疫苗', '预防脊髓灰质炎', 'https://baike.baidu.com/item/%E8%84%8A%E9%AB%93%E7%81%B0%E8%B4%A8%E7%82%8E%E7%96%AB%E8%8B%97/10929983?fromModule=search-result_lemma', EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000)
ON CONFLICT (name) DO NOTHING;
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 1, 60, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='脊髓灰质炎疫苗';
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 2, 90, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='脊髓灰质炎疫苗';
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 3, 120, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='脊髓灰质炎疫苗';
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 4, 1460, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='脊髓灰质炎疫苗';

INSERT INTO "vaccine" (vaccine_id, name, disease, link, ctime, utime)
VALUES (gen_random_uuid(), '百白破疫苗', '预防百日咳、白喉、破伤风', 'https://baike.baidu.com/item/%E7%99%BE%E7%99%BD%E7%A0%B4%E7%96%AB%E8%8B%97?fromModule=lemma_search-box', EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000)
ON CONFLICT (name) DO NOTHING;
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 1, 90, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='百白破疫苗';
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 2, 120, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='百白破疫苗';
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 3, 150, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='百白破疫苗';
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 4, 540, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='百白破疫苗';

INSERT INTO "vaccine" (vaccine_id, name, disease, link, ctime, utime)
VALUES (gen_random_uuid(), 'A群流脑多糖疫苗', '预防A群脑膜炎球菌流行性脑脊髓膜炎', 'https://baike.baidu.com/item/A%E7%BE%A4%E8%84%91%E8%86%9C%E7%82%8E%E7%90%83%E8%8F%8C%E5%A4%9A%E7%B3%96%E8%8F%8C%E8%8B%97/18866507?fromModule=search-result_lemma', EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000)
ON CONFLICT (name) DO NOTHING;
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 1, 180, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='A群流脑多糖疫苗';
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 2, 270, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='A群流脑多糖疫苗';

INSERT INTO "vaccine" (vaccine_id, name, disease, link, ctime, utime)
VALUES (gen_random_uuid(), '麻疹风疹联合疫苗', '预防麻疹、风疹', 'https://baike.baidu.com/item/%E9%BA%BB%E7%96%B9%E6%B5%81%E8%A1%8C%E6%80%A7%E8%85%AE%E8%85%BA%E7%82%8E%E9%A3%8E%E7%96%B9%E7%96%AB%E8%8B%97/6773488?fromModule=search-result_lemma', EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000)
ON CONFLICT (name) DO NOTHING;
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 1, 240, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='麻疹风疹联合疫苗';

INSERT INTO "vaccine" (vaccine_id, name, disease, link, ctime, utime)
VALUES (gen_random_uuid(), '乙型脑炎疫苗', '预防乙型脑炎', 'https://baike.baidu.com/item/%E4%B9%99%E5%9E%8B%E8%84%91%E7%82%8E%E5%87%8F%E6%AF%92%E7%96%AB%E8%8B%97/22251475?fromModule=search-result_lemma', EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000)
ON CONFLICT (name) DO NOTHING;
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 1, 240, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='乙型脑炎疫苗';
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 2, 730, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='乙型脑炎疫苗';

INSERT INTO "vaccine" (vaccine_id, name, disease, link, ctime, utime)
VALUES (gen_random_uuid(), '麻腮风疫苗', '预防麻疹、风疹、流行性腮腺炎', 'https://baike.baidu.com/item/%E9%BA%BB%E7%96%B9%E6%B5%81%E8%A1%8C%E6%80%A7%E8%85%AE%E8%85%BA%E7%82%8E%E9%A3%8E%E7%96%B9%E7%96%AB%E8%8B%97/6773488?fromModule=search-result_lemma', EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000)
ON CONFLICT (name) DO NOTHING;
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 1, 540, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='麻腮风疫苗';

INSERT INTO "vaccine" (vaccine_id, name, disease, link, ctime, utime)
VALUES (gen_random_uuid(), '甲型肝炎疫苗', '预防甲型肝炎', 'https://baike.baidu.com/item/%E7%94%B2%E5%9E%8B%E8%82%9D%E7%82%8E%E7%96%AB%E8%8B%97/55897827?fromModule=search-result_lemma', EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000)
ON CONFLICT (name) DO NOTHING;
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 1, 540, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='甲型肝炎疫苗';

INSERT INTO "vaccine" (vaccine_id, name, disease, link, ctime, utime)
VALUES (gen_random_uuid(), 'A群C群流脑多糖疫苗', '预防A群、C群脑膜炎球菌流行性脑脊髓膜炎', 'https://baike.baidu.com/item/A%E7%BE%A4%E6%B5%81%E8%84%91%E7%96%AB%E8%8B%97/12788994?fromModule=search-result_lemma', EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000)
ON CONFLICT (name) DO NOTHING;
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 1, 1095, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='A群C群流脑多糖疫苗';
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 2, 2190, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='A群C群流脑多糖疫苗';

INSERT INTO "vaccine" (vaccine_id, name, disease, link, ctime, utime)
VALUES (gen_random_uuid(), '白破疫苗', '预防白喉、破伤风（加强）', 'https://baike.baidu.com/item/%E7%99%BD%E7%A0%B4%E7%96%AB%E8%8B%97?fromModule=lemma_search-box', EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000)
ON CONFLICT (name) DO NOTHING;
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
SELECT gen_random_uuid(), vaccine_id, 1, 2190, EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000 FROM "vaccine" WHERE name='白破疫苗';
