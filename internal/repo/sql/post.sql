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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($2, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($2, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($2, '')::uuid
  ) AS is_collect
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($5, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($5, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($5, '')::uuid
  ) AS is_collect
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($3, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($3, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($3, '')::uuid
  ) AS is_collect
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($3, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($3, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($3, '')::uuid
  ) AS is_collect
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($4, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($4, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($4, '')::uuid
  ) AS is_collect
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($5, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($5, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($5, '')::uuid
  ) AS is_collect
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($4, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($4, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($4, '')::uuid
  ) AS is_collect
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($1::text, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($1::text, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($1::text, '')::uuid
  ) AS is_collect
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.author_id = NULLIF($1::text, '')::uuid AND p.status = 'published'
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($1::text, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($1::text, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($1::text, '')::uuid
  ) AS is_collect
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.author_id = NULLIF($1::text, '')::uuid AND p.status = 'published'
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($1::text, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($1::text, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($1::text, '')::uuid
  ) AS is_collect
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.author_id = NULLIF($1::text, '')::uuid AND p.status = 'draft'
ORDER BY p.ctime DESC
LIMIT $2 OFFSET $3;

-- name: ListMilestonesByAuthor :many
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($1::text, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($1::text, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($1::text, '')::uuid
  ) AS is_collect
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.author_id = NULLIF($1::text, '')::uuid AND p.status = 'milestone'
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($5, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($5, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($5, '')::uuid
  ) AS is_collect
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($4, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($4, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($4, '')::uuid
  ) AS is_collect
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($5, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($5, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($5, '')::uuid
  ) AS is_collect
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($5, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($5, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($5, '')::uuid
  ) AS is_collect
FROM "post" p
JOIN "post_tag" pt2 ON pt2.post_id = p.post_id
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE p.title ILIKE $1 AND pt2.tag_id = $2 AND p.status = 'published'
ORDER BY (p.like_count*3 + p.comment_count*5 + p.collect_count*4) DESC, p.ctime DESC
LIMIT $3 OFFSET $4;

-- name: ListFollowingByCtime :many
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($1, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($1, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($1, '')::uuid
  ) AS is_collect
FROM "post" p
JOIN "user_follow" f ON f.followee = p.author_id
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE f.follower = NULLIF($1, '')::uuid AND p.status = 'published'
ORDER BY p.ctime DESC
LIMIT $2 OFFSET $3;

-- name: ListFollowingByHot :many
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($1, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($1, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  EXISTS(
    SELECT 1 FROM "collection" c2
    WHERE c2.post_id = p.post_id AND c2.user_id = NULLIF($1, '')::uuid
  ) AS is_collect
FROM "post" p
JOIN "user_follow" f ON f.followee = p.author_id
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE f.follower = NULLIF($1, '')::uuid AND p.status = 'published'
ORDER BY (p.like_count*3 + p.comment_count*5 + p.collect_count*4) DESC, p.ctime DESC
LIMIT $2 OFFSET $3;

-- name: PublishPost :execrows
UPDATE "post"
SET status = 'published', utime = $3
WHERE post_id = $1 AND author_id = $2 AND status = 'draft';

-- name: UpdateDraftByOwner :execrows
UPDATE "post"
SET title = $3, content = $4, utime = $5
WHERE post_id = $1 AND author_id = $2 AND status = 'draft';

-- name: DeletePostTagsByPost :exec
DELETE FROM "post_tag" WHERE post_id = $1;

-- name: DeleteDraftByOwner :execrows
DELETE FROM "post"
WHERE post_id = $1 AND author_id = $2 AND status = 'draft';

-- name: GetPostStatusByID :one
SELECT status FROM "post" WHERE post_id = $1;

-- name: CreateComment :exec
INSERT INTO "comment" (
  comment_id, post_id, user_id, parent_id, content, status, like_count, reply_count, ctime, utime
) VALUES (
  $1, $2, $3, $4, $5, 'visible', 0, 0, $6, $7
);

-- name: GetCommentMinimal :one
SELECT post_id::text AS post_id, status
FROM "comment"
WHERE comment_id = $1;

-- name: IncPostCommentCount :execrows
UPDATE "post"
SET comment_count = comment_count + 1
WHERE post_id = $1;

-- name: IncCommentReplyCount :execrows
UPDATE "comment"
SET reply_count = reply_count + 1
WHERE comment_id = $1;

-- name: DecPostCommentCount :execrows
UPDATE "post"
SET comment_count = GREATEST(comment_count - 1, 0)
WHERE post_id = $1;

-- name: DecCommentReplyCount :execrows
UPDATE "comment"
SET reply_count = GREATEST(reply_count - 1, 0)
WHERE comment_id = $1;

-- name: ListPostCommentsByCtime :many
SELECT
  c.comment_id::text AS comment_id,
  c.user_id::text AS user_id,
  COALESCE(ua.username, '') AS username,
  COALESCE(ua.avatar, '') AS avatar,
  c.content,
  c.like_count,
  c.reply_count,
  c.ctime,
  c.utime,
  EXISTS(
    SELECT 1 FROM "comment_like" cl
    WHERE cl.comment_id = c.comment_id AND cl.user_id = NULLIF($4, '')::uuid
  ) AS has_liked
FROM "comment" c
LEFT JOIN "user_addition" ua ON ua.user_id = c.user_id
WHERE c.post_id = $1 AND c.parent_id IS NULL AND c.status = 'visible'
ORDER BY c.ctime DESC
LIMIT $2 OFFSET $3;

-- name: ListPostCommentsByHot :many
SELECT
  c.comment_id::text AS comment_id,
  c.user_id::text AS user_id,
  COALESCE(ua.username, '') AS username,
  COALESCE(ua.avatar, '') AS avatar,
  c.content,
  c.like_count,
  c.reply_count,
  c.ctime,
  c.utime,
  EXISTS(
    SELECT 1 FROM "comment_like" cl
    WHERE cl.comment_id = c.comment_id AND cl.user_id = NULLIF($4, '')::uuid
  ) AS has_liked
FROM "comment" c
LEFT JOIN "user_addition" ua ON ua.user_id = c.user_id
WHERE c.post_id = $1 AND c.parent_id IS NULL AND c.status = 'visible'
ORDER BY (c.like_count*3 + c.reply_count*5) DESC, c.ctime DESC
LIMIT $2 OFFSET $3;

-- name: ListCommentRepliesByCtime :many
SELECT
  c.comment_id::text AS comment_id,
  c.user_id::text AS user_id,
  COALESCE(ua.username, '') AS username,
  COALESCE(ua.avatar, '') AS avatar,
  c.content,
  c.like_count,
  (SELECT COUNT(1) FROM "comment" ch WHERE ch.parent_id = c.comment_id AND ch.status = 'visible') AS reply_count,
  c.ctime,
  c.utime,
  EXISTS(
    SELECT 1 FROM "comment_like" cl
    WHERE cl.comment_id = c.comment_id AND cl.user_id = NULLIF($4, '')::uuid
  ) AS has_liked
FROM "comment" c
LEFT JOIN "user_addition" ua ON ua.user_id = c.user_id
WHERE c.parent_id = $1 AND c.status = 'visible'
ORDER BY c.ctime DESC
LIMIT $2 OFFSET $3;

-- name: ListCommentRepliesByHot :many
SELECT
  c.comment_id::text AS comment_id,
  c.user_id::text AS user_id,
  COALESCE(ua.username, '') AS username,
  COALESCE(ua.avatar, '') AS avatar,
  c.content,
  c.like_count,
  (SELECT COUNT(1) FROM "comment" ch WHERE ch.parent_id = c.comment_id AND ch.status = 'visible') AS reply_count,
  c.ctime,
  c.utime,
  EXISTS(
    SELECT 1 FROM "comment_like" cl
    WHERE cl.comment_id = c.comment_id AND cl.user_id = NULLIF($4, '')::uuid
  ) AS has_liked
FROM "comment" c
LEFT JOIN "user_addition" ua ON ua.user_id = c.user_id
WHERE c.parent_id = $1 AND c.status = 'visible'
ORDER BY c.like_count DESC, c.ctime DESC
LIMIT $2 OFFSET $3;

-- name: GetCommentMetaByID :one
SELECT
  post_id::text AS post_id,
  parent_id,
  status,
  user_id::text AS user_id
FROM "comment"
WHERE comment_id = $1;

-- name: DeleteCommentVisibleByOwner :execrows
UPDATE "comment"
SET status = 'deleted', utime = $2
WHERE comment_id = $1 AND user_id = $3 AND status = 'visible';

-- name: UpdateCommentContentByOwner :execrows
UPDATE "comment"
SET content = $2, utime = $3
WHERE comment_id = $1 AND user_id = $4 AND status = 'visible';

-- name: CreateCommentLike :execrows
INSERT INTO "comment_like" (user_id, comment_id, ctime, utime)
VALUES ($1, $2, $3, $3)
ON CONFLICT (user_id, comment_id) DO NOTHING;

-- name: DeleteCommentLike :execrows
DELETE FROM "comment_like"
WHERE user_id = $1 AND comment_id = $2;

-- name: IncCommentLikeCount :execrows
UPDATE "comment"
SET like_count = like_count + 1
WHERE comment_id = $1 AND status = 'visible';

-- name: DecCommentLikeCount :execrows
UPDATE "comment"
SET like_count = GREATEST(like_count - 1, 0)
WHERE comment_id = $1;

-- name: CreateCollection :execrows
INSERT INTO "collection" (collection_id, user_id, post_id, ctime, utime)
VALUES ($1, $2, $3, $4, $4)
ON CONFLICT (user_id, post_id) DO NOTHING;

-- name: DeleteCollection :execrows
DELETE FROM "collection"
WHERE user_id = $1 AND post_id = $2;

-- name: IncPostCollectCount :execrows
UPDATE "post"
SET collect_count = collect_count + 1
WHERE post_id = $1 AND status = 'published';

-- name: DecPostCollectCount :execrows
UPDATE "post"
SET collect_count = GREATEST(collect_count - 1, 0)
WHERE post_id = $1;

-- name: ListCollectionsByCtime :many
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($1, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($1, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  TRUE AS is_collect
FROM "collection" c
JOIN "post" p ON p.post_id = c.post_id
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE c.user_id = NULLIF($1, '')::uuid
ORDER BY c.ctime DESC
LIMIT $2 OFFSET $3;

-- name: ListCollectionsByHot :many
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
  ), '{}') AS tags,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld
    WHERE ld.post_id = p.post_id AND ld.user_id = NULLIF($1, '')::uuid AND ld.type = 'like'
  ) AS is_like,
  EXISTS(
    SELECT 1 FROM "like_dislike" ld2
    WHERE ld2.post_id = p.post_id AND ld2.user_id = NULLIF($1, '')::uuid AND ld2.type = 'dislike'
  ) AS is_dislike,
  TRUE AS is_collect
FROM "collection" c
JOIN "post" p ON p.post_id = c.post_id
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
WHERE c.user_id = NULLIF($1, '')::uuid AND p.status = 'published'
ORDER BY (p.like_count*3 + p.comment_count*5 + p.collect_count*4) DESC, c.ctime DESC
LIMIT $2 OFFSET $3;
-- name: CreatePostLike :execrows
INSERT INTO "like_dislike" (user_id, post_id, type, ctime, utime)
VALUES ($1, $2, 'like', $3, $3)
ON CONFLICT (user_id, post_id) DO UPDATE
SET type = 'like', utime = $3
WHERE like_dislike.type <> 'like';


-- name: DeletePostLike :execrows
DELETE FROM "like_dislike"
WHERE user_id = $1 AND post_id = $2 AND type = 'like';

-- name: IncPostLikeCount :execrows
UPDATE "post"
SET like_count = like_count + 1
WHERE post_id = $1 AND status = 'published';

-- name: DecPostLikeCount :execrows
UPDATE "post"
SET like_count = GREATEST(like_count - 1, 0)
WHERE post_id = $1;

-- admin
-- name: CreateTag :one
INSERT INTO "tag"(tag_id, tag_name, description, ctime, utime)
VALUES ($1, $2, $3, $4, $4)
RETURNING tag_id::text AS tag_id, tag_name, COALESCE(description, '') AS description;

-- admin
-- name: DeletePostTagsByTagID :execrows
DELETE FROM "post_tag" WHERE tag_id = $1;

-- admin
-- name: DeleteTagByID :execrows
DELETE FROM "tag" WHERE tag_id = $1;

-- admin
-- name: ListTags :many
SELECT 
  tag_id::text AS tag_id,
  tag_name,
  COALESCE(description, '') AS description
FROM "tag"
WHERE ($1 = '' OR tag_name ILIKE '%' || $1 || '%')
ORDER BY ctime DESC
LIMIT $2 OFFSET $3;
