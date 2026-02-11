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
  COALESCE(b.birthday, 0) AS birthday,
  COALESCE(array_agg(t.tag_name) FILTER (WHERE t.tag_name IS NOT NULL), '{}') AS tags
FROM "post" p
LEFT JOIN "user_addition" ua ON ua.user_id = p.author_id
LEFT JOIN LATERAL (
  SELECT birthday FROM "baby"
  WHERE user_id = p.author_id
  ORDER BY ctime DESC
  LIMIT 1
) b ON true
LEFT JOIN "post_tag" pt ON pt.post_id = p.post_id
LEFT JOIN "tag" t ON t.tag_id = pt.tag_id
WHERE p.post_id = $1
GROUP BY
  p.post_id, p.author_id, ua.username, ua.avatar, ua.province, ua.city,
  p.title, p.content, p.status, p.like_count, p.dislike_count, p.collect_count, p.comment_count,
  p.cover, p.ctime, p.utime, b.birthday;

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
