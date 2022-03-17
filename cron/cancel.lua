local job = KEYS[1]
local hostname = ARGV[1]

local host = redis.call('GET', job .. ':host')
if host == nil then
    return -2
end

if host ~= hostname then
    return -1
end
return redis.call('DEL', job .. ':host')

