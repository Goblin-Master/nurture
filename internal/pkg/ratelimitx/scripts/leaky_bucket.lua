local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local capacity = limit

local level = tonumber(redis.call("HGET", key, "level"))
local ts = tonumber(redis.call("HGET", key, "ts"))

if level == nil then
    level = 0
end
if ts == nil then
    ts = now
end

local elapsed = now - ts
if elapsed < 0 then
    elapsed = 0
end

level = math.max(0, level - elapsed * capacity / window)

local allowed = 0
if level + 1 <= capacity then
    level = level + 1
    allowed = 1
end

local remaining = math.floor(capacity - level)
if remaining < 0 then
    remaining = 0
end

redis.call("HSET", key, "level", level, "ts", now)
redis.call("PEXPIRE", key, window * 2)

return {allowed, remaining}
