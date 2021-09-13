directorymonitor = {}
local common = require("common")
-- This method is to detect whether in the directory file md5 has changed,and the parameter is the directory path
function directorymonitor.change(dir)
    local exit = false
    while not exit do
        if file_exist(dir)  then
            common.watcher(dir, {2, 4, 8, 16})
        end
        sc_sleep(5)
    end
end

-- This method is to detect whether in the directory file md5 has changed,and the parameter is the directory path
function directorymonitor.add(dir)
    local exit = false
    while not exit do
        if file_exist(dir)  then
            common.watcher(dir, {1})
        end
        sc_sleep(5)
    end
end

function directorymonitor.del(dir)
    local cache_key = dir

    local old = get_cache(cache_key)
    if not old then
        local currents = ls(dir)

        set_cache(cache_key, currents)
        return
    end

    local content=''
    local currents = ls(dir)
    for i,v in ipairs(old) do
        local exist=false
        for ii,vv in ipairs(currents) do
            if vv["filename"] == v["filename"] and vv["path"] == v["path"] then
                exist = true
                break
            end
        end
        if not exist then
            if content ~= '' then content=content.."; " end
            content = content..string.format("%s %s", v["path"], v["mode"])
        end
    end

    if content ~= '' then
        trigger({Content=content})
        set_cache(cache_key, currents)
    elseif #old ~= #currents then
        set_cache(cache_key, currents)
    end
end


function directorymonitor.priv_change(dir)
    local cache_key = dir

    local old = get_cache(cache_key)
    if old == nil then
        local currents = ls(dir)
        set_cache(cache_key, currents)
        return
    end
    local content=''

    local currents = ls(dir)
    for i,v in ipairs(old) do
        local exist=false
        local oldmode = ''
        for ii,vv in ipairs(currents) do
            if vv["mode"] ~= v["mode"] and vv["path"] == v["path"] then
                exist = true
                oldmode = vv['mode']
                break
            end
        end
        if  exist then
            if content ~= '' then content=content.."; " end
            content = content..string.format("%s mode %s change %s ", v["path"], v["mode"], oldmode)
        end
    end

    if content ~= ''
    then
        trigger({Content=content})
        set_cache(cache_key, currents)
    elseif #old ~= #currents then
        set_cache(cache_key, currents)
    end
end

return directorymonitor
