directorymonitor = {}
local common = require("common")
-- This method is to detect whether in the directory file md5 has changed,and the parameter is the directory path
function directorymonitor.change(dir)
    --local dirkey = dir
    --local filelist = ls(dir)
    --local olddir = get_cache(dirkey)
    --if olddir == nil then
    --    for ii,value in ipairs(filelist) do
    --        local file = ''
    --        if  string.sub(value['mode'],0,1) == '-'
    --        then
    --            file = value['path']
    --            local cache_key=file
    --            local hashval = file_hash(file)
    --            set_cache(cache_key, hashval)
    --        end
    --    end
    --    set_cache(dirkey, filelist)
    --    return
    --end
    --local content=''
    --for ii,value in ipairs(olddir) do
    --    local file = ''
    --    if  string.sub(value['mode'],0,1) == '-'
    --    then
    --        file = value['path']
    --        local cache_key=file
    --        local hashval = file_hash(file)
    --        local exist=false
    --        local old = get_cache(cache_key)
    --
    --        if old == hashval then
    --            exist = true
    --        end
    --        if not exist then
    --            if content ~= '' then content=content.."; " end
    --            content = content..string.format("%s", file)
    --            set_cache(cache_key, hashval)
    --        end
    --    end
    --end
    --if content ~= ''
    --then
    --    trigger({Content=content})
    --end
    local exit = false
    while not exit do
        if file_exist(dir)  then
            common.watcher(dir, 2)
        end
        sc_sleep(5)
    end
end

-- This method is to detect whether in the directory file md5 has changed,and the parameter is the directory path
function directorymonitor.add(dir)
    --local cache_key = dir
    --local old = get_cache(cache_key)
    --if old == nil then
    --    local currents = ls(dir)
    --
    --    set_cache(cache_key, currents)
    --    return
    --end
    --local content=''
    --local currents = ls(dir)
    --for i,v in ipairs(currents) do
    --    local exist=false
    --    for ii,vv in ipairs(old) do
    --        if vv["filename"] == v["filename"] and vv["path"] == v["path"]  then
    --            --print(vv['name'])
    --            exist = true
    --            break
    --        end
    --    end
    --    if not exist then
    --        if content ~= '' then content=content.."; " end
    --        content = content..string.format("%s %s", v["path"], v["mode"])
    --    end
    --end
    --
    --if content ~= '' then
    --    trigger({Content=content})
    --    set_cache(cache_key, currents)
    --elseif #old ~= #currents then
    --    set_cache(cache_key, currents)
    --end
    local exit = false
    while not exit do
        if file_exist(dir)  then
            common.watcher(dir, 1)
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
