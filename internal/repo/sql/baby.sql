-- name: ListBabiesByUserID :many
SELECT baby_id::text AS baby_id, name, avatar FROM "baby"
WHERE user_id = $1
ORDER BY ctime DESC;

-- name: CreateBaby :exec
INSERT INTO "baby" (baby_id, user_id, name, gender, birthday, avatar, ctime, utime)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetBabyByIDAndUser :one
SELECT * FROM "baby" WHERE baby_id = $1 AND user_id = $2 LIMIT 1;

-- name: CreateBabyGrowthRecord :exec
INSERT INTO "baby_growth_record" (record_id, baby_id, user_id, record_time, height, weight, head_circumference, remark, ctime, utime)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetLatestGrowthByBabyIDAndUser :one
SELECT * FROM "baby_growth_record"
WHERE baby_id = $1 AND user_id = $2
ORDER BY record_time DESC
LIMIT 1;

-- 列出所有剂次（用于初始化宝宝疫苗记录）
-- name: ListAllDoses :many
SELECT d.dose_id, d.vaccine_id, v.name, v.disease, d.recommend_age_days, v.link, v.ctime, v.utime, d.dose_number
FROM "vaccine_dose" d
JOIN "vaccine" v ON v.vaccine_id = d.vaccine_id
ORDER BY v.name ASC, d.dose_number ASC;

-- name: CreateBabyVaccineRecord :exec
INSERT INTO "baby_vaccine_record" (record_id, baby_id, dose_id, due_time, status, actual_time, ctime, utime)
VALUES ($1, $2, $3, $4, 'not_given', NULL, $5, $6);

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
