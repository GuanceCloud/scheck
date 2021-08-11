local valuemonitor = require("valuemonitor")


local function check()
    local cache_key = "kernel_version"
    local old = get_cache(cache_key)
    if old == nil then
        local current = kernel_info()['version']
        set_cache(cache_key, current)
        return
    end
    local current =  kernel_info()['version']
    if old ~= current then
        trigger({Content=current})
        set_cache(cache_key, current)
    end

end
check()
