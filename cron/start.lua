local job = KEYS[1]
local hostname = ARGV[1]
local expire = tonumber(ARGV[2])

local host = redis.call('GET', job .. ':host')
if host ~= false and host ~= hostname then
    return -2
end

redis.call('SET', job .. ':host', hostname)
redis.call('EXPIRE', job .. ':host', expire)

local expectedNext = redis.call('GET', job .. ':next')
if expectedNext == false then
    return -1
end

return expectedNext


