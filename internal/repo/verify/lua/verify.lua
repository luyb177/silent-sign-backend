-- verify.lua
-- ARGV[1] = user-submitted code
-- Returns: 1 = matched + deleted, 0 = mismatch, nil = key not found

local stored = redis.call("GET", KEYS[1])

if not stored then
    return nil
end

if stored ~= ARGV[1] then
    return 0
end

redis.call("DEL", KEYS[1])
return 1