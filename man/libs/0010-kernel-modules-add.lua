local system = require("system")
local cache = require("cache")
local function check()
    local cache_key="kernel-modules"
    local currents=system.kernel_modules()

    local old=cache.get_cache(cache_key)
    if not old then
        cache.set_cache(cache_key, currents)
        return
    end

    local content=''

    for i,v in ipairs(currents) do
        local exist=false
        for ii,vv in ipairs(old) do
            if vv["name"] == v["name"] then
                 --print(vv['name'])
                exist = true
                break
            end
        end
        if not exist then
            if content ~= '' then content=content.."; " end
            content = content..string.format("%s", v["name"])
        end
    end

    if content ~= '' then
        trigger({Content=content})
        cache.set_cache(cache_key, currents)
    end

end
check()