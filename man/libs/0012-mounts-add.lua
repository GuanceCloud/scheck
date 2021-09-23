
local function check()
    local cache_key="mounts"
    local currents=mounts()

    local old=get_cache(cache_key)

    if old == nil then
        set_cache(cache_key, currents)
        return
    end

    local content=''

    for i,v in ipairs(currents) do
        local exist=false
        for ii,vv in ipairs(old) do
            if vv["path"] == v["path"] then
--                  print(vv['name'])
                exist = true
                break
            end
        end
        if not exist then
            if content ~= '' then content=content.."; " end
            content = content..string.format("%s mount %s", v["device"], v["path"])
        end
    end

    if content ~= '' then
        trigger({Content=content})
        set_cache(cache_key, currents)
    end

end
check()
