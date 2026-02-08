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
