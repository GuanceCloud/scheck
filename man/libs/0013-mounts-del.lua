
local function check()
    local cache_key="mounts"
    local currents=mounts()

    local old=get_cache(cache_key)
    if not old then
        set_cache(cache_key, currents)
        return
    end

    local content=''

    for i,v in ipairs(old) do
        local exist=false
        for ii,vv in ipairs(currents) do
            if vv["path"] == v["path"] then
                -- print(vv['path'])
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
        -- print('-----')
        -- print(content)
        trigger({Content=content})
        set_cache(cache_key, currents)
    end

end
check()
