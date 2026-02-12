-- name: GetPostDetail :one
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt
    JOIN "tag" t ON t.tag_id = pt.tag_id
    WHERE pt.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.post_id = $1;

-- name: CreatePost :exec
INSERT INTO "post" (
  post_id, author_id, title, content, status,
  like_count, dislike_count, collect_count, comment_count,
  cover, ctime, utime
) VALUES (
  $1, $2, $3, $4, $5,
  0, 0, 0, 0,
  $6, $7, $8
);

-- name: AddPostTag :exec
INSERT INTO "post_tag" (post_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListHomePosts :many
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt
    JOIN "tag" t ON t.tag_id = pt.tag_id
    WHERE pt.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.status = $1
ORDER BY 
  CASE WHEN $2 = 'hot' THEN (p.like_count*3 + p.comment_count*5 + p.collect_count*4) ELSE p.ctime END DESC,
  p.ctime DESC
LIMIT $3 OFFSET $4;

-- name: ListHomeByHot :many
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt
    JOIN "tag" t ON t.tag_id = pt.tag_id
    WHERE pt.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.status = 'published'
ORDER BY (p.like_count*3 + p.comment_count*5 + p.collect_count*4) DESC, p.ctime DESC
LIMIT $1 OFFSET $2;

-- name: ListHomeByCtime :many
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt
    JOIN "tag" t ON t.tag_id = pt.tag_id
    WHERE pt.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.status = 'published'
ORDER BY p.ctime DESC
LIMIT $1 OFFSET $2;

-- name: ListHomeByRandom :many
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt
    JOIN "tag" t ON t.tag_id = pt.tag_id
    WHERE pt.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.status = 'published'
ORDER BY md5(p.post_id::text || ($1)::text)
LIMIT $2 OFFSET $3;
-- name: ListPostsByTag :many
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt2
    JOIN "tag" t ON t.tag_id = pt2.tag_id
    WHERE pt2.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
JOIN "post_tag" pt ON pt.post_id = p.post_id
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE pt.tag_id = $1 AND p.status = $2
ORDER BY p.ctime DESC
LIMIT $3 OFFSET $4;

-- name: ListPostsByTagHot :many
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt2
    JOIN "tag" t ON t.tag_id = pt2.tag_id
    WHERE pt2.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
JOIN "post_tag" pt ON pt.post_id = p.post_id
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE pt.tag_id = $1 AND p.status = 'published'
ORDER BY (p.like_count*3 + p.comment_count*5 + p.collect_count*4) DESC, p.ctime DESC
LIMIT $2 OFFSET $3;

-- name: ListPostsByAuthor :many
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt2
    JOIN "tag" t ON t.tag_id = pt2.tag_id
    WHERE pt2.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.author_id = $1 AND p.status = 'published'
ORDER BY p.ctime DESC
LIMIT $2 OFFSET $3;

-- name: ListPostsByAuthorHot :many
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt2
    JOIN "tag" t ON t.tag_id = pt2.tag_id
    WHERE pt2.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.author_id = $1 AND p.status = 'published'
ORDER BY (p.like_count*3 + p.comment_count*5 + p.collect_count*4) DESC, p.ctime DESC
LIMIT $2 OFFSET $3;
-- name: ListDraftsByAuthor :many
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt2
    JOIN "tag" t ON t.tag_id = pt2.tag_id
    WHERE pt2.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.author_id = $1 AND p.status = 'draft'
ORDER BY p.ctime DESC
LIMIT $2 OFFSET $3;

-- name: SearchPosts :many
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt
    JOIN "tag" t ON t.tag_id = pt.tag_id
    WHERE pt.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.title ILIKE $1 AND p.status = $2
ORDER BY p.ctime DESC
LIMIT $3 OFFSET $4;

-- name: SearchPostsByTitleHot :many
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt
    JOIN "tag" t ON t.tag_id = pt.tag_id
    WHERE pt.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.title ILIKE $1 AND p.status = 'published'
ORDER BY (p.like_count*3 + p.comment_count*5 + p.collect_count*4) DESC, p.ctime DESC
LIMIT $2 OFFSET $3;

-- name: SearchPostsByTitleAndTagCtime :many
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt
    JOIN "tag" t ON t.tag_id = pt.tag_id
    WHERE pt.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
JOIN "post_tag" pt2 ON pt2.post_id = p.post_id
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.title ILIKE $1 AND pt2.tag_id = $2 AND p.status = 'published'
ORDER BY p.ctime DESC
LIMIT $3 OFFSET $4;

-- name: SearchPostsByTitleAndTagHot :many
SELECT
  p.post_id::text AS post_id,
  p.author_id::text AS author_id,
  COALESCE(ua.username, '') AS author_name,
  COALESCE(ua.avatar, '') AS author_avatar,
  COALESCE(ua.province, '') AS author_province,
  COALESCE(ua.city, '') AS author_city,
  p.title, p.content, p.status,
  p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime,
  COALESCE((
    SELECT b.birthday
    FROM "baby" b
    WHERE b.user_id = p.author_id
    ORDER BY b.ctime DESC
    LIMIT 1
  ), 0) AS birthday,
  COALESCE((
    SELECT array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL)
    FROM "post_tag" pt
    JOIN "tag" t ON t.tag_id = pt.tag_id
    WHERE pt.post_id = p.post_id
  ), '{}') AS tags
FROM "post" p
JOIN "post_tag" pt2 ON pt2.post_id = p.post_id
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.title ILIKE $1 AND pt2.tag_id = $2 AND p.status = 'published'
ORDER BY (p.like_count*3 + p.comment_count*5 + p.collect_count*4) DESC, p.ctime DESC
LIMIT $3 OFFSET $4;
