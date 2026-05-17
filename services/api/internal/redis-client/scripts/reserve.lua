-- reserve.lua
-- Reserve all remaining daily quota for a live session.
-- Returns the reserved amount.
--
-- KEYS[1] = user_quota:{user}:{date}
-- ARGV[1] = max_daily
-- ARGV[2] = ttl_seconds

local key = KEYS[1]
local max_daily = tonumber(ARGV[1])
local ttl_seconds = tonumber(ARGV[2])

if not max_daily or max_daily <= 0 then
    return 0
end

if not ttl_seconds or ttl_seconds <= 0 then
    ttl_seconds = 1
end

local remaining = redis.call('GET', key)

if not remaining then
    -- First request today: full daily quota is available.
    -- Reserve all of it by setting remaining to 0.
    redis.call('SET', key, 0, 'EX', ttl_seconds)
    return max_daily
end

remaining = tonumber(remaining)

if not remaining or remaining <= 0 then
    redis.call('EXPIRE', key, ttl_seconds)
    return 0
end

-- Reserve all currently remaining quota.
redis.call('SET', key, 0, 'EX', ttl_seconds)
return remaining