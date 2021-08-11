
local function check()

    local cache_key = "hostname"
    local old = get_cache(cache_key)
    if old == nil then
        local current = hostname()
        set_cache(cache_key, current)
        return
    end
    local current =  hostname()
    if old ~= current then
        trigger({Content=current})
        set_cache(cache_key, current)
    end
end
check()
