-- release.lua
local key = KEYS[1]
local reserve_amount = tonumber(ARGV[1])   -- what you reserved earlier
local actual_usage = tonumber(ARGV[2])    -- what was really used
local max_daily = tonumber(ARGV[3])
local ttl_seconds = tonumber(ARGV[4])

local remaining = redis.call('GET', key)
if not remaining then
    -- Key disappeared (shouldn’t happen if reservation held it)
    -- Rebuild: treat as if full quota was there, then apply actual usage.
    -- (Simplified: just set to max_daily - actual_usage, but beware of race)
    -- For brevity, assume reservation always keeps key alive.
    if max_daily < actual_usage then
        redis.call('SET', key, 0, 'EX', ttl_seconds)
    else
        redis.call('SET', key, max_daily - actual_usage, 'EX', ttl_seconds)
    end
    return 1
end

remaining = tonumber(remaining)
local unused = reserve_amount - actual_usage
if unused > 0 then
    redis.call('INCRBY', key, unused)
end
-- No need to decrease if actual_usage > reserve (shouldn't happen by design)
-- Keep TTL refreshed
redis.call('EXPIRE', key, ttl_seconds)
return 1