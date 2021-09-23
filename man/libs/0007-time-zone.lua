local function check()
    local cache_key = "time_zone"
    local old = get_cache(cache_key)
    if old == nil then
        local current = time_zone()
        set_cache(cache_key, current)
        return
    end
    local current = time_zone()
    if old ~= current then
        trigger({Content=current})
        set_cache(cache_key, current)
    end

end
check()
