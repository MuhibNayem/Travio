-- enqueue.lua
-- KEYS[1] = queue_key (Sorted Set)
-- KEYS[2] = user_entry_key (String/Hash)
-- ARGV[1] = user_id
-- ARGV[2] = current_timestamp (score)
-- ARGV[3] = max_active_users (optional limit check)
-- ARGV[4] = entry_json_data
-- ARGV[5] = entry_ttl

local queue_key = KEYS[1]
local entry_key = KEYS[2]
local user_id = ARGV[1]
local score = tonumber(ARGV[2])
local max_active = tonumber(ARGV[3])
local entry_data = ARGV[4]
local ttl = tonumber(ARGV[5])

-- Check if user is already in queue
if redis.call("ZSCORE", queue_key, user_id) then
    -- Already queued, return current rank
    local rank = redis.call("ZRANK", queue_key, user_id)
    return {1, rank + 1} -- 1 = existing, +1 for 1-based rank
end

-- Add to queue
redis.call("ZADD", queue_key, score, user_id)
redis.call("SET", entry_key, entry_data, "EX", ttl)

-- Get new rank
local rank = redis.call("ZRANK", queue_key, user_id)

return {0, rank + 1} -- 0 = new entry
