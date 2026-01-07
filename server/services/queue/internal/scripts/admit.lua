-- admit.lua
-- KEYS[1] = queue_waiting_key (Sorted Set)
-- KEYS[2] = queue_stats_key (Hash)
-- ARGV[1] = count (number of users to admit)
-- ARGV[2] = token_ttl (seconds)
-- ARGV[3] = event_id

local waiting_key = KEYS[1]
local stats_key = KEYS[2]
local count = tonumber(ARGV[1])
local token_ttl = tonumber(ARGV[2])
local event_id = ARGV[3]

-- Get top N users from the waiting line
local users = redis.call("ZRANGE", waiting_key, 0, count - 1)

if #users == 0 then
    return {}
end

local admitted = {}

for _, user_id in ipairs(users) do
    -- 1. Remove from waiting queue
    redis.call("ZREM", waiting_key, user_id)
    
    -- 2. Update their specific entry key to mark as READY
    local entry_key = string.format("queue:%s:entry:%s", event_id, user_id)
    local entry_json = redis.call("GET", entry_key)
    
    if entry_json then
        -- Strict parsing using cjson, no assumptions about string formatting
        local entry = cjson.decode(entry_json)
        
        entry.status = "ready"
        -- We can also update expires_at if needed, but time handling in Lua is tricky (os.time() is not allowed)
        -- The service layer can handle specific expiry logic or we just rely on TTL.
        -- We will just update status here.
        
        local new_json = cjson.encode(entry)
        
        -- Extend TTL for the admission window as they are now active
        redis.call("SET", entry_key, new_json, "EX", token_ttl)
    end
    
    table.insert(admitted, user_id)
end

-- 3. Update global stats
redis.call("HINCRBY", stats_key, "admitted", #admitted)

return admitted
