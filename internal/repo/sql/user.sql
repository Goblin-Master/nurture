-- name: GetUserByAccountAndPassword :one
SELECT * FROM "user_base"
WHERE account = $1 AND password = $2 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM "user_base"
WHERE email = $1 LIMIT 1;

-- name: CreateUser :exec
WITH ins AS (
  INSERT INTO "user_base" (
    user_id, ctime, utime, account, password, email, username, gender, role
  ) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, 1
  )
  RETURNING user_id, username, gender, ctime, utime
)
INSERT INTO "user_addition" (
  user_id, username, gender, phone, occupation, birthday, avatar, province, city, ctime, utime
) SELECT
  user_id, username, gender, '', '', 0, '', '', '', ctime, utime
FROM ins;

-- name: UpdatePasswordByEmail :execrows
UPDATE "user_base"
SET password = $2
WHERE email = $1;

-- name: UpdateAvatarByUserID :execrows
UPDATE "user_addition"
SET avatar = $2
WHERE user_id = $1;

-- name: UpdateGenderByUserID :execrows
UPDATE "user_addition"
SET gender = $2, utime = $3
WHERE user_id = $1;

-- name: UpdateBaseGenderByUserID :execrows
UPDATE "user_base"
SET gender = $2, utime = $3
WHERE user_id = $1;

-- name: CreatePartner :exec
INSERT INTO "user_partner"(father, mother, ctime, utime)
VALUES ($1, $2, $3, $3)
ON CONFLICT (father, mother) DO NOTHING;

-- name: GetPartnerByUserID :one
SELECT father::text AS father, mother::text AS mother
FROM "user_partner"
WHERE father = $1 OR mother = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM "user_base" WHERE user_id = $1 LIMIT 1;

-- name: UpdateUserAdditionByUserID :execrows
UPDATE "user_addition"
SET
  occupation = COALESCE(NULLIF($2, ''), occupation),
  phone      = COALESCE(NULLIF($3, ''), phone),
  province   = COALESCE(NULLIF($4, ''), province),
  city       = COALESCE(NULLIF($5, ''), city),
  avatar     = COALESCE(NULLIF($6, ''), avatar),
  birthday   = COALESCE(NULLIF($7::BIGINT, -1), birthday),
  utime      = $8
WHERE user_id = $1;
