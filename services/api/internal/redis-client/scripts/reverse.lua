-- reserve.lua
local key = KEYS[1]                    -- user_quota:{user}:{date}
local reserve_amount = tonumber(ARGV[1])
local max_daily = tonumber(ARGV[2])
local ttl_seconds = tonumber(ARGV[3])

local remaining = redis.call('GET', key)

if not remaining then
    -- First request today: initialise full quota
    if max_daily < reserve_amount then
        return 0   -- not enough even at full quota
    end
    redis.call('SET', key, max_daily - reserve_amount, 'EX', ttl_seconds)
    return 1
end

remaining = tonumber(remaining)
if remaining >= reserve_amount then
    redis.call('DECRBY', key, reserve_amount)
    redis.call('EXPIRE', key, ttl_seconds)   -- refresh TTL
    return 1
else
    return 0   -- reservation denied
end