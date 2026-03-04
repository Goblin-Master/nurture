-- name: ListBabiesByUserID :many
SELECT baby_id::text AS baby_id, name, avatar FROM "baby"
WHERE user_id = $1
ORDER BY ctime DESC;

-- name: CreateBaby :exec
INSERT INTO "baby" (baby_id, user_id, name, gender, birthday, avatar, ctime, utime)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (baby_id, user_id) DO NOTHING;

-- name: GetBabyByIDAndUser :one
SELECT * FROM "baby" WHERE baby_id = $1 AND user_id = $2 LIMIT 1;

-- name: CreateBabyGrowthRecord :exec
INSERT INTO "baby_growth_record" (record_id, baby_id, record_time, height, weight, head_circumference, remark, ctime, utime, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetLatestGrowthByBabyID :one
SELECT * FROM "baby_growth_record"
WHERE baby_id = $1
ORDER BY record_time DESC
LIMIT 1;

-- name: GetLatestNonNullGrowthValuesByBaby :one
SELECT
  (SELECT bgr.height FROM "baby_growth_record" bgr WHERE bgr.baby_id = $1 AND bgr.height IS NOT NULL ORDER BY bgr.record_time DESC LIMIT 1) AS height,
  (SELECT bgr.weight FROM "baby_growth_record" bgr WHERE bgr.baby_id = $1 AND bgr.weight IS NOT NULL ORDER BY bgr.record_time DESC LIMIT 1) AS weight,
  (SELECT bgr.head_circumference FROM "baby_growth_record" bgr WHERE bgr.baby_id = $1 AND bgr.head_circumference IS NOT NULL ORDER BY bgr.record_time DESC LIMIT 1) AS head_circumference;

-- name: UpdateGrowthByRecordID :exec
UPDATE "baby_growth_record"
SET record_time = $2,
    height = $3,
    weight = $4,
    head_circumference = $5,
    remark = COALESCE(NULLIF($6, ''), remark),
    updated_by = $7,
    utime = $8
WHERE record_id = $1;

-- 列出所有剂次（用于初始化宝宝疫苗记录）
-- name: ListAllDoses :many
SELECT d.dose_id, d.vaccine_id, v.name, v.disease, d.recommend_age_days, v.link, v.ctime, v.utime, d.dose_number
FROM "vaccine_dose" d
JOIN "vaccine" v ON v.vaccine_id = d.vaccine_id
ORDER BY v.name ASC, d.dose_number ASC;

-- name: CreateBabyVaccineRecord :exec
INSERT INTO "baby_vaccine_record" (record_id, baby_id, dose_id, due_time, status, actual_time, ctime, utime)
VALUES ($1, $2, $3, $4, 'not_given', NULL, $5, $6);

-- 管理员：创建疫苗
-- name: CreateVaccine :one
INSERT INTO "vaccine" (vaccine_id, name, disease, link, ctime, utime)
VALUES ($1, $2, $3, $4, $5, $5)
RETURNING vaccine_id::text AS vaccine_id;

-- 管理员：为疫苗新增剂次
-- name: CreateVaccineDose :one
INSERT INTO "vaccine_dose" (dose_id, vaccine_id, dose_number, recommend_age_days, ctime, utime)
VALUES ($1, $2, $3, $4, $5, $5)
RETURNING dose_id::text AS dose_id, dose_number, recommend_age_days;

-- 管理员：为新剂次初始化所有宝宝的接种记录（未接种）
-- name: InitBabyVaccineRecordsForDose :execrows
INSERT INTO "baby_vaccine_record" (record_id, baby_id, dose_id, due_time, status, actual_time, ctime, utime)
SELECT gen_random_uuid(), b.baby_id, $1, b.birthday + (d.recommend_age_days * 24 * 3600 * 1000), 'not_given', NULL, $2, $2
FROM "baby" b
JOIN "vaccine_dose" d ON d.dose_id = $1
ON CONFLICT (baby_id, dose_id) DO NOTHING;

-- name: ListVaccineRecordsByBabyID :many
SELECT 
  d.dose_id::text AS dose_id,
  v.vaccine_id::text AS vaccine_id,
  v.name,
  v.disease,
  v.link,
  d.dose_number,
  bvr.due_time,
  bvr.status,
  COALESCE(bvr.actual_time, 0) AS actual_time
FROM "baby_vaccine_record" bvr
JOIN "vaccine_dose" d ON d.dose_id = bvr.dose_id
JOIN "vaccine" v ON v.vaccine_id = d.vaccine_id
WHERE bvr.baby_id = $1
ORDER BY v.name ASC, d.dose_number ASC;

-- name: UpdateVaccineStatusGiven :execrows
UPDATE "baby_vaccine_record"
SET status = 'given', actual_time = $3, utime = $4
WHERE baby_id = $1 AND dose_id = $2;

-- name: UpdateVaccineStatusNotGiven :execrows
UPDATE "baby_vaccine_record"
SET status = 'not_given', actual_time = NULL, utime = $3
WHERE baby_id = $1 AND dose_id = $2;

-- name: UploadBabyPhotos :many
WITH lnk AS (SELECT unnest($2::text[]) AS link)
INSERT INTO "baby_photo" (photo_id, baby_id, link, ctime, utime)
SELECT gen_random_uuid(), $1, lnk.link, $3, $3 FROM lnk
RETURNING photo_id::text AS photo_id, link, ctime;

-- name: DeleteBabyPhotos :execrows
DELETE FROM "baby_photo"
WHERE baby_id = $1 AND photo_id = ANY($2::uuid[]);

-- name: ListBabyPhotos :many
SELECT photo_id::text AS photo_id, link, ctime
FROM "baby_photo"
WHERE baby_id = $1
ORDER BY ctime DESC
LIMIT $2 OFFSET $3;

-- name: GetGrowthByBabyIDBetween :one
SELECT * FROM "baby_growth_record"
WHERE baby_id = $1 AND record_time BETWEEN $2 AND $3
ORDER BY record_time DESC
LIMIT 1;

-- name: ListHeightCurveByBabyIDBetween :many
SELECT record_time, height
FROM "baby_growth_record"
WHERE baby_id = $1 AND record_time BETWEEN $2 AND $3 AND height IS NOT NULL
ORDER BY record_time ASC;

-- name: ListWeightCurveByBabyIDBetween :many
SELECT record_time, weight
FROM "baby_growth_record"
WHERE baby_id = $1 AND record_time BETWEEN $2 AND $3 AND weight IS NOT NULL
ORDER BY record_time ASC;

-- name: ListHeadCircumferenceCurveByBabyIDBetween :many
SELECT record_time, head_circumference
FROM "baby_growth_record"
WHERE baby_id = $1 AND record_time BETWEEN $2 AND $3 AND head_circumference IS NOT NULL
ORDER BY record_time ASC;

-- name: StartSleep :one
WITH ins AS (
  INSERT INTO "daily_sleep" (
    sleep_id, baby_id, user_id, start_time, end_time, duration,
    source, status, ctime, utime
  )
  VALUES (
    gen_random_uuid(), $1, $2, EXTRACT(EPOCH FROM NOW())*1000, NULL, NULL,
    'timer', 'running', EXTRACT(EPOCH FROM NOW())*1000, EXTRACT(EPOCH FROM NOW())*1000
  )
  ON CONFLICT DO NOTHING
  RETURNING sleep_id, baby_id, user_id, start_time
)
SELECT sleep_id::text AS sleep_id, baby_id::text AS baby_id, user_id::text AS user_id, start_time
FROM ins
UNION ALL
SELECT sleep_id::text, baby_id::text, user_id::text, start_time
FROM "daily_sleep"
WHERE baby_id = $1 AND user_id = $2 AND status = 'running'
LIMIT 1;

-- name: StopSleep :one
UPDATE "daily_sleep"
SET
  end_time = EXTRACT(EPOCH FROM NOW())*1000,
  duration = EXTRACT(EPOCH FROM NOW())*1000 - start_time,
  status   = 'finished',
  utime    = EXTRACT(EPOCH FROM NOW())*1000
WHERE sleep_id = $1
RETURNING sleep_id::text AS sleep_id, baby_id::text AS baby_id, user_id::text AS user_id, start_time, end_time, duration;

-- name: ForceStopSleepWithCap :one
UPDATE "daily_sleep"
SET
  end_time = start_time + $2,
  duration = $2,
  status   = 'finished',
  utime    = EXTRACT(EPOCH FROM NOW())*1000
WHERE sleep_id = $1 AND status = 'running'
RETURNING sleep_id::text AS sleep_id, start_time, end_time, duration;

-- name: GetActiveSleep :one
SELECT sleep_id::text AS sleep_id, start_time
FROM "daily_sleep"
WHERE baby_id = $1 AND user_id = $2 AND status = 'running'
LIMIT 1;

-- name: CreateFeeding :one
INSERT INTO "daily_feeding" (
  feeding_id, baby_id, user_id, feed_time, feed_type, remark, ctime, utime
) VALUES (
  gen_random_uuid(), $1, $2, $3, $4, COALESCE(NULLIF($5, ''), NULL), $6, $6
) RETURNING feeding_id::text AS feeding_id, feed_time, feed_type, COALESCE(remark, '') AS remark;

-- name: UpdateFeeding :one
UPDATE "daily_feeding"
SET
  feed_type = $3,
  feed_time = $4,
  remark    = COALESCE(NULLIF($5, ''), NULL),
  utime     = $6
WHERE baby_id = $1 AND feeding_id = $2
RETURNING feeding_id::text AS feeding_id, feed_time, feed_type, COALESCE(remark, '') AS remark;

-- name: ListSleepByBabyBetween :many
SELECT sleep_id::text AS session_id, start_time, end_time, duration
FROM "daily_sleep"
WHERE baby_id = $1
  AND status = 'finished'
  AND start_time <= $3
  AND end_time >= $2
ORDER BY start_time ASC;

-- name: ListFeedingByBabyBetween :many
SELECT feeding_id::text AS feeding_id, feed_time, feed_type, COALESCE(remark, '') AS remark
FROM "daily_feeding"
WHERE baby_id = $1 AND feed_time BETWEEN $2 AND $3
ORDER BY feed_time ASC;

-- name: CreateDiaper :one
INSERT INTO "daily_diaper" (
  diaper_id, baby_id, user_id, change_time, diaper_type,
  pee_color, poop_color, poop_consistency, remark, ctime, utime
) VALUES (
  gen_random_uuid(), $1, $2, $3, $4,
  COALESCE(NULLIF($5, ''), NULL),
  COALESCE(NULLIF($6, ''), NULL),
  COALESCE(NULLIF($7, ''), NULL),
  COALESCE(NULLIF($8, ''), NULL), $9, $9
) RETURNING diaper_id::text AS diaper_id, change_time, diaper_type,
  COALESCE(pee_color, '') AS pee_color,
  COALESCE(poop_color, '') AS poop_color,
  COALESCE(poop_consistency, '') AS poop_consistency,
  COALESCE(remark, '') AS remark;

-- name: UpdateDiaper :one
UPDATE "daily_diaper"
SET
  diaper_type      = $3,
  change_time      = $4,
  pee_color        = COALESCE(NULLIF($5, ''), NULL),
  poop_color       = COALESCE(NULLIF($6, ''), NULL),
  poop_consistency = COALESCE(NULLIF($7, ''), NULL),
  remark           = COALESCE(NULLIF($8, ''), NULL),
  utime            = $9
WHERE baby_id = $1 AND diaper_id = $2
RETURNING diaper_id::text AS diaper_id, change_time, diaper_type,
  COALESCE(pee_color, '') AS pee_color,
  COALESCE(poop_color, '') AS poop_color,
  COALESCE(poop_consistency, '') AS poop_consistency,
  COALESCE(remark, '') AS remark;

-- name: GetDiaperByBabyBetween :one
SELECT diaper_id::text AS diaper_id, change_time, diaper_type,
  COALESCE(pee_color, '') AS pee_color,
  COALESCE(poop_color, '') AS poop_color,
  COALESCE(poop_consistency, '') AS poop_consistency,
  COALESCE(remark, '') AS remark
FROM "daily_diaper"
WHERE baby_id = $1 AND change_time BETWEEN $2 AND $3
ORDER BY change_time DESC
LIMIT 1;

-- name: ListDiaperByBabyBetween :many
SELECT diaper_id::text AS diaper_id, change_time, diaper_type,
  COALESCE(pee_color, '') AS pee_color,
  COALESCE(poop_color, '') AS poop_color,
  COALESCE(poop_consistency, '') AS poop_consistency,
  COALESCE(remark, '') AS remark
FROM "daily_diaper"
WHERE baby_id = $1 AND change_time BETWEEN $2 AND $3
ORDER BY change_time ASC;

-- name: GetDailyStatsByBaby :one
WITH feed_stats AS (
  SELECT COUNT(*) AS c
  FROM "daily_feeding" f
  WHERE f.baby_id = $1 AND f.feed_time BETWEEN $2 AND $3
),
sleep_stats AS (
  SELECT COALESCE(SUM(LEAST(s.end_time, $3) - GREATEST(s.start_time, $2)), 0)::bigint AS d
  FROM "daily_sleep" s
  WHERE s.baby_id = $1 AND s.status = 'finished' AND s.start_time <= $3 AND s.end_time >= $2
),
diaper_stats AS (
  SELECT COUNT(*) AS c
  FROM "daily_diaper" d
  WHERE d.baby_id = $1 AND d.change_time BETWEEN $2 AND $3
)
SELECT feed_stats.c AS feeding_count, sleep_stats.d AS sleep_duration_ms, diaper_stats.c AS diaper_count
FROM feed_stats, sleep_stats, diaper_stats;
