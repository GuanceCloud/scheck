
local function check()
    local cache_key="kernel-modules"
    local currents=kernel_modules()

    local old=get_cache(cache_key)
    if not old then
        set_cache(cache_key, currents)
        return
    end

    local content=''

    for i,v in ipairs(old) do
        local exist=false
        for ii,vv in ipairs(currents) do
            if vv["name"] == v["name"] then
                 -- print(vv['name'])
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
        -- print('-----')
        -- print(content)
        trigger({Content=content})
        set_cache(cache_key, currents)
    end

end
check()