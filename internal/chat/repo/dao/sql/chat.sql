-- name: CreateChatGroup :exec
WITH g AS (
  INSERT INTO "chat_group" (
    group_id, owner_id, name, avatar, description,
    member_limit, member_count, ctime, utime
  ) VALUES (
    $1, $2, $3, $4, $5,
    $6, 1, $7, $7
  )
  RETURNING group_id, owner_id, ctime, utime
)
INSERT INTO "chat_group_member" (
  group_id, user_id, role, ctime, utime
) SELECT
  group_id, owner_id, 'owner', ctime, utime
FROM g;

-- name: GetChatGroupByID :one
SELECT * FROM "chat_group"
WHERE group_id = $1
LIMIT 1;

-- name: GetChatGroupProfile :one
SELECT
  g.group_id::text AS group_id,
  g.owner_id::text AS owner_user_id,
  COALESCE(ua.username, '') AS owner_username,
  COALESCE(ua.avatar, '') AS owner_avatar,
  g.name,
  g.avatar,
  g.description,
  g.member_limit,
  g.member_count,
  g.ctime,
  g.utime
FROM "chat_group" g
LEFT JOIN "user_addition" ua ON ua.user_id = g.owner_id
WHERE g.group_id = $1
LIMIT 1;

-- name: LockChatGroupByID :one
SELECT * FROM "chat_group"
WHERE group_id = $1
LIMIT 1
FOR UPDATE;

-- name: IncChatGroupMemberCount :execrows
UPDATE "chat_group"
SET member_count = member_count + 1,
    utime = $2
WHERE group_id = $1;

-- name: DecChatGroupMemberCount :execrows
UPDATE "chat_group"
SET member_count = member_count - 1,
    utime = $2
WHERE group_id = $1;

-- name: CreateChatGroupMember :exec
INSERT INTO "chat_group_member" (
  group_id, user_id, role, ctime, utime
) VALUES (
  $1, $2, $3, $4, $4
)
ON CONFLICT (group_id, user_id) DO NOTHING;

-- name: DeleteChatGroupMember :execrows
DELETE FROM "chat_group_member"
WHERE group_id = $1 AND user_id = $2;

-- name: UpdateChatGroupMemberLastSeenTime :execrows
UPDATE "chat_group_member"
SET last_seen_time = $3,
    utime = $3
WHERE group_id = $1 AND user_id = $2;

-- name: GetChatGroupMemberRole :one
SELECT role
FROM "chat_group_member"
WHERE group_id = $1 AND user_id = $2
LIMIT 1;

-- name: TransferChatGroupOwner :execrows
UPDATE "chat_group_member"
SET role = CASE
  WHEN user_id = $2 THEN 'owner'
  WHEN user_id = $3 THEN $4
  ELSE role
END,
utime = $5
WHERE group_id = $1 AND user_id IN ($2, $3);

-- name: UpdateChatGroupOwnerID :execrows
UPDATE "chat_group"
SET owner_id = $2,
    utime = $3
WHERE group_id = $1;

-- name: DeleteChatGroupMessagesByGroupID :execrows
DELETE FROM "chat_group_message"
WHERE group_id = $1;

-- name: DeleteChatGroupMembersByGroupID :execrows
DELETE FROM "chat_group_member"
WHERE group_id = $1;

-- name: DeleteChatGroupByID :execrows
DELETE FROM "chat_group"
WHERE group_id = $1;

-- name: ListMyChatGroups :many
SELECT
  gm.group_id::text AS group_id,
  g.name,
  g.avatar,
  g.description,
  g.member_limit,
  g.member_count,
  gm.role,
  g.ctime,
  g.utime,
  COALESCE(um.unread_count, 0) AS unread_count,
  COALESCE(lm.from_user_id::text, '') AS last_message_from_user_id,
  COALESCE(ualm.username, '') AS last_message_from_username,
  COALESCE(lm.type, '') AS last_message_type,
  COALESCE(lm.content, '') AS last_message_content,
  COALESCE(lm.ctime, 0) AS last_message_time
FROM "chat_group_member" gm
JOIN "chat_group" g ON g.group_id = gm.group_id
LEFT JOIN LATERAL (
  SELECT COUNT(1) AS unread_count
  FROM "chat_group_message" m2
  WHERE m2.group_id = gm.group_id
    AND m2.ctime > gm.last_seen_time
    AND m2.from_user_id <> gm.user_id
) um ON TRUE
LEFT JOIN LATERAL (
  SELECT m.from_user_id, m.type, m.content, m.ctime
  FROM "chat_group_message" m
  WHERE m.group_id = gm.group_id
  ORDER BY m.ctime DESC, m.message_id DESC
  LIMIT 1
) lm ON TRUE
LEFT JOIN "user_addition" ualm ON ualm.user_id = lm.from_user_id
WHERE gm.user_id = $1
ORDER BY COALESCE(lm.ctime, 0) DESC, g.ctime DESC;

-- name: ListDiscoverChatGroupsFirst :many
SELECT
  g.group_id::text AS group_id,
  g.name,
  g.avatar,
  g.member_count,
  g.member_limit,
  md5($2::text || g.group_id::text) AS sort_key
FROM "chat_group" g
WHERE g.member_count < g.member_limit
  AND NOT EXISTS (
    SELECT 1
    FROM "chat_group_member" gm
    WHERE gm.group_id = g.group_id
      AND gm.user_id = $1
  )
ORDER BY sort_key ASC, g.group_id ASC
LIMIT $3;

-- name: ListDiscoverChatGroupsAfter :many
SELECT
  g.group_id::text AS group_id,
  g.name,
  g.avatar,
  g.member_count,
  g.member_limit,
  md5($2::text || g.group_id::text) AS sort_key
FROM "chat_group" g
WHERE g.member_count < g.member_limit
  AND NOT EXISTS (
    SELECT 1
    FROM "chat_group_member" gm
    WHERE gm.group_id = g.group_id
      AND gm.user_id = $1
  )
  AND (
    md5($2::text || g.group_id::text) > $3::text OR
    (md5($2::text || g.group_id::text) = $3::text AND g.group_id > $4)
  )
ORDER BY sort_key ASC, g.group_id ASC
LIMIT $5;

-- name: SearchChatGroupsByName :many
SELECT
  g.group_id::text AS group_id,
  g.name,
  g.avatar,
  g.member_count,
  g.member_limit
FROM "chat_group" g
WHERE g.name ILIKE ('%' || $1 || '%')
ORDER BY g.member_count DESC, g.ctime DESC
LIMIT $2;

-- name: ListChatGroupMembersWithProfile :many
SELECT
  gm.user_id::text AS user_id,
  COALESCE(ua.username, '') AS username,
  COALESCE(ua.avatar, '') AS avatar,
  gm.role,
  gm.ctime AS join_time
FROM "chat_group_member" gm
LEFT JOIN "user_addition" ua ON ua.user_id = gm.user_id
WHERE gm.group_id = $1
ORDER BY
  CASE gm.role WHEN 'owner' THEN 1 WHEN 'admin' THEN 2 ELSE 3 END,
  gm.ctime ASC
LIMIT $2 OFFSET $3;

-- name: ListChatGroupMembersPreviewWithProfile :many
SELECT
  gm.user_id::text AS user_id,
  COALESCE(ua.username, '') AS username,
  COALESCE(ua.avatar, '') AS avatar,
  gm.role,
  gm.ctime AS join_time
FROM "chat_group_member" gm
LEFT JOIN "user_addition" ua ON ua.user_id = gm.user_id
WHERE gm.group_id = $1
ORDER BY
  CASE gm.role WHEN 'owner' THEN 1 WHEN 'admin' THEN 2 ELSE 3 END,
  gm.ctime ASC
LIMIT $2;

-- name: CreateChatGroupMessage :execrows
INSERT INTO "chat_group_message" (
  message_id, group_id, from_user_id, type, content, ctime, utime
) VALUES (
  $1, $2, $3, $4, $5, $6, $6
)
ON CONFLICT (group_id, message_id) DO NOTHING;

-- name: CreateChatDirectMessage :execrows
INSERT INTO "chat_direct_message" (
  message_id, from_user_id, to_user_id, type, content, ctime, utime
) VALUES (
  $1, $2, $3, $4, $5, $6, $6
)
ON CONFLICT (from_user_id, to_user_id, message_id) DO NOTHING;

-- name: UpsertChatDirectSeen :exec
INSERT INTO "chat_direct_seen" (
  user_id, partner_user_id, last_seen_time, ctime, utime
) VALUES (
  $1, $2, $3, $4, $4
)
ON CONFLICT (user_id, partner_user_id) DO UPDATE
SET last_seen_time = GREATEST("chat_direct_seen".last_seen_time, EXCLUDED.last_seen_time),
    utime = EXCLUDED.utime;

-- name: CreateChatEventOutbox :execrows
INSERT INTO "chat_event_outbox" (
  event_id, routing_key, payload, status,
  attempts, next_retry_at, published_at, ctime, utime
) VALUES (
  $1, $2, $3, 'pending',
  0, 0, 0, $4, $4
)
ON CONFLICT (event_id) DO NOTHING;

-- name: ClaimPendingChatEventOutbox :many
WITH picked AS (
  SELECT e.id
  FROM "chat_event_outbox" e
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
UPDATE "chat_event_outbox" o
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

-- name: MarkChatEventOutboxPublished :execrows
UPDATE "chat_event_outbox"
SET status = 'published',
    published_at = $2,
    utime = $2
WHERE id = $1
  AND status = 'publishing';

-- name: MarkChatEventOutboxFailed :execrows
UPDATE "chat_event_outbox"
SET status = CASE WHEN attempts + 1 >= $3 THEN 'failed' ELSE 'pending' END,
    attempts = attempts + 1,
    next_retry_at = CASE WHEN attempts + 1 >= $3 THEN 0 ELSE $2 END,
    utime = $4
WHERE id = $1
  AND status = 'publishing';

-- name: ListChatGroupMessagesLatest :many
SELECT
  message_id::text AS message_id,
  group_id::text AS group_id,
  from_user_id::text AS from_user_id,
  type,
  content,
  ctime
FROM "chat_group_message"
WHERE group_id = $1
ORDER BY ctime DESC, message_id DESC
LIMIT $2;

-- name: ListChatGroupMessagesBefore :many
SELECT
  message_id::text AS message_id,
  group_id::text AS group_id,
  from_user_id::text AS from_user_id,
  type,
  content,
  ctime
FROM "chat_group_message"
WHERE group_id = $1
  AND (
    ctime < $2 OR
    (ctime = $2 AND message_id < $3)
  )
ORDER BY ctime DESC, message_id DESC
LIMIT $4;

-- name: ListChatGroupMessagesAfter :many
SELECT
  message_id::text AS message_id,
  group_id::text AS group_id,
  from_user_id::text AS from_user_id,
  type,
  content,
  ctime
FROM "chat_group_message"
WHERE group_id = $1
  AND (
    ctime > $2 OR
    (ctime = $2 AND message_id > $3)
)
ORDER BY ctime ASC, message_id ASC
LIMIT $4;

-- name: ListChatDirectMessagesLatest :many
SELECT
  message_id::text AS message_id,
  from_user_id::text AS from_user_id,
  to_user_id::text AS to_user_id,
  type,
  content,
  ctime
FROM "chat_direct_message"
WHERE (
  from_user_id = $1 AND to_user_id = $2
) OR (
  from_user_id = $2 AND to_user_id = $1
)
ORDER BY ctime DESC, message_id DESC
LIMIT $3;

-- name: ListChatDirectMessagesBefore :many
SELECT
  message_id::text AS message_id,
  from_user_id::text AS from_user_id,
  to_user_id::text AS to_user_id,
  type,
  content,
  ctime
FROM "chat_direct_message"
WHERE (
  (
    from_user_id = $1 AND to_user_id = $2
  ) OR (
    from_user_id = $2 AND to_user_id = $1
  )
)
  AND (
    ctime < $3 OR
    (ctime = $3 AND message_id < $4)
  )
ORDER BY ctime DESC, message_id DESC
LIMIT $5;

-- name: ListChatDirectMessagesAfter :many
SELECT
  message_id::text AS message_id,
  from_user_id::text AS from_user_id,
  to_user_id::text AS to_user_id,
  type,
  content,
  ctime
FROM "chat_direct_message"
WHERE (
  (
    from_user_id = $1 AND to_user_id = $2
  ) OR (
    from_user_id = $2 AND to_user_id = $1
  )
)
  AND (
    ctime > $3 OR
    (ctime = $3 AND message_id > $4)
  )
ORDER BY ctime ASC, message_id ASC
LIMIT $5;
