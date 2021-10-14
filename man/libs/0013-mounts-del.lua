local cache = require("cache")
local function check()
    local cache_key="mounts"
    local system = require("system")
    local currents=system.mounts()

    local old=cache.get_cache(cache_key)
    if not old then
        cache.set_cache(cache_key, currents)
        return
    end

    local content=''

    for i,v in ipairs(old) do
        local exist=false
        for ii,vv in ipairs(currents) do
            if vv["path"] == v["path"] then
                exist = true
                break
            end
        end
        if not exist then
            if content ~= '' then content=content.."; " end
            content = content..string.format("%s umount %s", v["device"], v["path"])
        end
    end

    if content ~= '' then
        trigger({Content=content})
        cache.set_cache(cache_key, currents)
    end

end
check()
