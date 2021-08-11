
local function check()
    local cache_key = "password_status"
    local old = get_cache(cache_key)

    local count = 0
    local content = ''

    if not old then
        set_cache(cache_key, shadow())
        return
    end
    local currents = shadow()
    for i,v in ipairs(old) do
        local exist=true
        for ii,vv in ipairs(currents) do
            if  v['username'] == vv['username']
            then
                if v['password_status'] ~= vv['password_status'] then
                    exist = false
                    break
                end
            end
        end
        if not exist then
            if content ~= '' then content=content.."; " end
            count = count + 1
            content = content..string.format("%s", v["username"])
        end
    end
    if content ~= '' then
        trigger({Content=content})
        set_cache(cache_key, currents)
    end


end
check()
