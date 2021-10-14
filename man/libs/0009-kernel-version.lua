local valuemonitor = require("valuemonitor")
local system = require("system")
local cache = require("cache")
local function check()
    local cache_key = "kernel_version"
    local old = cache.get_cache(cache_key)
    if old == nil then
        local current = system.kernel_info()['version']
        cache.set_cache(cache_key, current)
        return
    end
    local current =   system.kernel_info()['version']
    if old ~= current then
        trigger({Content=current})
        cache.set_cache(cache_key, current)
    end

end
check()
