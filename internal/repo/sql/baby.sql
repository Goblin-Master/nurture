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

-- name: ListAllVaccines :many
SELECT vaccine_id, name, disease, recommend_age_days, link, ctime, utime
FROM "vaccine"
ORDER BY recommend_age_days ASC;

-- name: CreateBabyVaccineRecord :exec
INSERT INTO "baby_vaccine_record" (record_id, baby_id, vaccine_id, due_time, status, actual_time, ctime, utime)
VALUES ($1, $2, $3, $4, 'not_given', NULL, $5, $6);

-- name: ListVaccineRecordsByBabyID :many
SELECT v.vaccine_id::text AS vaccine_id, v.name, v.disease, bvr.due_time, bvr.status, COALESCE(bvr.actual_time, 0) AS actual_time
FROM "baby_vaccine_record" bvr
JOIN "vaccine" v ON v.vaccine_id = bvr.vaccine_id
WHERE bvr.baby_id = $1
ORDER BY v.recommend_age_days ASC;

-- name: UpdateVaccineStatusGiven :execrows
UPDATE "baby_vaccine_record"
SET status = 'given', actual_time = $3, utime = $4
WHERE baby_id = $1 AND vaccine_id = $2;

-- name: UpdateVaccineStatusNotGiven :execrows
UPDATE "baby_vaccine_record"
SET status = 'not_given', actual_time = NULL, utime = $3
WHERE baby_id = $1 AND vaccine_id = $2;
