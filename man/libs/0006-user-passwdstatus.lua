local system = require("system")
local cache = require("cache")
local function check()
    local cache_key = "password_status"
    local old = cache.get_cache(cache_key)

    local count = 0
    local content = ''

    if not old then
        cache.set_cache(cache_key, system.shadow())
        return
    end
    local currents = system.shadow()
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
        cache.set_cache(cache_key, currents)
    end


end
check()
