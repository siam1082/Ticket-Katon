package service

// acquireLockLua checks if ALL requested seat keys are free.
// If all are free, it sets all keys with the user_id and TTL.
// If any key already exists, it aborts immediately and returns 0.
const acquireLockLua = `
local ttl = tonumber(ARGV[1])
local user_id = ARGV[2]

-- Step 1: Verify none of the seats are currently locked
for i, key in ipairs(KEYS) do
    if redis.call("EXISTS", key) == 1 then
        return 0 -- Conflict: at least one seat is already locked
    end
end

-- Step 2: Acquire lock on all seats atomically
for i, key in ipairs(KEYS) do
    redis.call("SET", key, user_id, "EX", ttl)
end

return 1 -- Success: all seats locked
`

// releaseLockLua verifies that the caller owns the lock before deleting it.
const releaseLockLua = `
local user_id = ARGV[1]
local released_count = 0

for i, key in ipairs(KEYS) do
    if redis.call("GET", key) == user_id then
        redis.call("DEL", key)
        released_count = released_count + 1
    end
end

return released_count
`