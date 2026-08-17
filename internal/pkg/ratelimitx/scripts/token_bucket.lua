local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local capacity = limit

local tokens = tonumber(redis.call("HGET", key, "tokens"))
local ts = tonumber(redis.call("HGET", key, "ts"))

if tokens == nil then
    tokens = capacity
end
if ts == nil then
    ts = now
end

local elapsed = now - ts
if elapsed < 0 then
    elapsed = 0
end

tokens = math.min(capacity, tokens + elapsed * capacity / window)

local allowed = 0
if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

redis.call("HSET", key, "tokens", tokens, "ts", now)
redis.call("PEXPIRE", key, window * 2)

return {allowed, math.floor(tokens)}
