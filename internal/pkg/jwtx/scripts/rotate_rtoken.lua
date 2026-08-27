local old_active_key = KEYS[1]
local new_active_key = KEYS[2]
local old_used_key = KEYS[3]

local new_session = ARGV[1]
local active_ttl_ms = tonumber(ARGV[2])
local used_ttl_ms = tonumber(ARGV[3])

if redis.call("EXISTS", old_active_key) == 0 then
    if redis.call("EXISTS", old_used_key) == 1 then
        return -1
    end
    return 0
end

redis.call("DEL", old_active_key)
redis.call("SET", new_active_key, new_session, "PX", active_ttl_ms)

if used_ttl_ms > 0 then
    redis.call("SET", old_used_key, "1", "PX", used_ttl_ms)
end

return 1
