local cache = require("cache")
local function check()
    local system = require("system")
    local cache_key = "time_zone"
    local old = cache.get_cache(cache_key)
    if old == nil then
        local current = system.time_zone()
        cache.set_cache(cache_key, current)
        return
    end
    local current = system.time_zone()
    if old ~= current then
        trigger({Content=current})
        cache.set_cache(cache_key, current)
    end

end
check()
