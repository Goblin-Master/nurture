-- name: GetUserByAccountAndPassword :one
SELECT * FROM "user"
WHERE account = $1 AND password = $2 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM "user"
WHERE email = $1 LIMIT 1;

-- name: CreateUser :exec
INSERT INTO "user" (
  user_id, ctime, utime, account, password, email, username, avatar, role
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: UpdatePasswordByEmail :execrows
UPDATE "user"
SET password = $2
WHERE email = $1;

-- name: UpdateAvatarByUserID :execrows
UPDATE "user"
SET avatar = $2
WHERE user_id = $1;