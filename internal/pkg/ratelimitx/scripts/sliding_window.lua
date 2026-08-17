local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local member = ARGV[4]

redis.call("ZREMRANGEBYSCORE", key, 0, now - window)

local count = redis.call("ZCARD", key)
if count >= limit then
    redis.call("PEXPIRE", key, window)
    return {0, 0}
end

redis.call("ZADD", key, now, member)
redis.call("PEXPIRE", key, window)

return {1, limit - count - 1}
