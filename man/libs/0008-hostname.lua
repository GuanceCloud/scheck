local cache = require("cache")
local function check()
    local system = require("system")
    local cache_key = "hostname"
    local old = cache.get_cache(cache_key)
    if old == nil then
        local current = system.hostname()
        cache.set_cache(cache_key, current)
        return
    end
    local current =  system.hostname()
    if old ~= current then
        trigger({Content=current})
        cache.set_cache(cache_key, current)
    end
end
check()
