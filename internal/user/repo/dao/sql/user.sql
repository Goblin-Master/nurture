-- name: GetUserByAccount :one
SELECT * FROM "user_base"
WHERE account = $1
LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM "user_base"
WHERE email = $1::varchar
LIMIT 1;

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

-- name: UpdatePasswordByUserID :execrows
UPDATE "user_base"
SET password = $2, utime = $3
WHERE user_id = $1;

-- name: BindEmailByUserID :execrows
UPDATE "user_base"
SET email = $2, utime = $3
WHERE user_id = $1;

-- name: GetMyProfile :one
SELECT 
  ub.user_id::text AS user_id,
  ub.account,
  COALESCE(ub.email, '') AS email,
  ub.username,
  ub.gender,
  ua.avatar,
  ua.phone,
  ua.occupation,
  ua.birthday,
  ua.province,
  ua.city,
  ub.ctime,
  ub.utime
FROM "user_base" ub
JOIN "user_addition" ua ON ua.user_id = ub.user_id
WHERE ub.user_id = $1
LIMIT 1;

-- name: CreateFollow :exec
INSERT INTO "user_follow"(follower, followee, ctime, utime)
VALUES ($1, $2, $3, $3)
ON CONFLICT (follower, followee) DO NOTHING;

-- name: DeleteFollow :execrows
DELETE FROM "user_follow"
WHERE follower = $1 AND followee = $2;

-- name: IsFollowing :one
SELECT EXISTS(
  SELECT 1 FROM "user_follow"
  WHERE follower = $1 AND followee = $2
) AS is_following;

-- name: IsPhoneUsed :one
SELECT EXISTS(
  SELECT 1 FROM "user_addition"
  WHERE phone = $1 AND phone <> '' AND user_id <> $2
) AS is_phone_used;

-- name: ListFollowingByUserID :many
SELECT 
  ub.user_id::text AS user_id,
  ub.username,
  ua.avatar,
  uf.ctime AS follow_time
FROM "user_follow" uf
JOIN "user_base" ub ON ub.user_id = uf.followee
JOIN "user_addition" ua ON ua.user_id = uf.followee
WHERE uf.follower = $1
ORDER BY uf.ctime DESC
LIMIT $2 OFFSET $3;

-- name: ListFollowersByUserID :many
SELECT 
  ub.user_id::text AS user_id,
  ub.username,
  ua.avatar,
  uf.ctime AS follow_time
FROM "user_follow" uf
JOIN "user_base" ub ON ub.user_id = uf.follower
JOIN "user_addition" ua ON ua.user_id = uf.follower
WHERE uf.followee = $1
ORDER BY uf.ctime DESC
LIMIT $2 OFFSET $3;

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

-- name: CreateUserEventOutbox :execrows
INSERT INTO "user_event_outbox" (
  event_id, routing_key, payload, status,
  attempts, next_retry_at, published_at, ctime, utime
) VALUES (
  $1, $2, $3, 'pending',
  0, 0, 0, $4, $4
)
ON CONFLICT (event_id) DO NOTHING;

-- name: ClaimPendingUserEventOutbox :many
WITH picked AS (
  SELECT e.id
  FROM "user_event_outbox" e
  WHERE (
    e.status = 'pending'
    AND e.next_retry_at <= sqlc.arg(retry_before)
  ) OR (
    e.status = 'publishing'
    AND e.utime <= sqlc.arg(stale_before)
  )
  ORDER BY e.id ASC
  LIMIT sqlc.arg(claim_limit)
  FOR UPDATE SKIP LOCKED
)
UPDATE "user_event_outbox" o
SET status = 'publishing',
    utime = sqlc.arg(claimed_at)
FROM picked
WHERE o.id = picked.id
RETURNING
  o.id,
  o.event_id,
  o.routing_key,
  o.payload,
  o.attempts,
  o.ctime;

-- name: MarkUserEventOutboxPublished :execrows
UPDATE "user_event_outbox"
SET status = 'published',
    published_at = $2,
    utime = $2
WHERE id = $1
  AND status = 'publishing';

-- name: MarkUserEventOutboxFailed :execrows
UPDATE "user_event_outbox"
SET status = CASE WHEN attempts + 1 >= $3 THEN 'failed' ELSE 'pending' END,
    attempts = attempts + 1,
    next_retry_at = CASE WHEN attempts + 1 >= $3 THEN 0 ELSE $2 END,
    utime = $4
WHERE id = $1
  AND status = 'publishing';

-- name: GetPartnerByUserID :one
SELECT father::text AS father, mother::text AS mother
FROM "user_partner"
WHERE father = $1 OR mother = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM "user_base"
WHERE user_id = $1
LIMIT 1;

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

-- admin
-- name: AdminListUsers :many
SELECT 
  ub.user_id::text AS user_id,
  ub.username,
  COALESCE(ua.avatar, '') AS avatar
FROM "user_base" ub
LEFT JOIN "user_addition" ua ON ua.user_id = ub.user_id
WHERE ($1 = '' OR ub.username ILIKE '%' || $1 || '%')
ORDER BY ub.ctime DESC
LIMIT $2 OFFSET $3;

-- admin
-- name: AdminUpdateUserRole :execrows
UPDATE "user_base"
SET role = $2, utime = $3
WHERE user_id = $1;
