-- release.lua
-- Release unused reserved quota after a live session.
--
-- KEYS[1] = user_quota:{user}:{date}
-- ARGV[1] = reserved_amount
-- ARGV[2] = actual_usage
-- ARGV[3] = max_daily
-- ARGV[4] = ttl_seconds

local key = KEYS[1]
local reserved_amount = tonumber(ARGV[1])
local actual_usage = tonumber(ARGV[2])
local max_daily = tonumber(ARGV[3])
local ttl_seconds = tonumber(ARGV[4])

if not reserved_amount or reserved_amount < 0 then
    reserved_amount = 0
end

if not actual_usage or actual_usage < 0 then
    actual_usage = 0
end

if not max_daily or max_daily < 0 then
    max_daily = 0
end

if not ttl_seconds or ttl_seconds <= 0 then
    ttl_seconds = 1
end

local remaining = redis.call('GET', key)

if not remaining then
    -- If key disappeared, rebuild based on actual usage.
    local new_remaining = max_daily - actual_usage

    if new_remaining < 0 then
        new_remaining = 0
    end

    redis.call('SET', key, new_remaining, 'EX', ttl_seconds)
    return 1
end

remaining = tonumber(remaining)

if not remaining or remaining < 0 then
    remaining = 0
end

local unused = reserved_amount - actual_usage

if unused < 0 then
    unused = 0
end

local new_remaining = remaining + unused

if new_remaining > max_daily then
    new_remaining = max_daily
end

redis.call('SET', key, new_remaining, 'EX', ttl_seconds)

return 1