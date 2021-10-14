local cache = require("cache")
local filemonitor = require("filemonitor")
local function check(file)
    local is_install_docker = cache.get_global_cache('install_docker')
    if is_install_docker == nil or not is_install_docker then
        return
    end
    filemonitor.priv_root_ownership(file)
end

file = '/usr/lib/systemd/system/docker.service'
check(file)
--
--local function is_root(file)
--    local mode = sc_file.file_info(file)
--    local uid = mode['uid']
--    local gid = mode['gid']
--    local  res = uid + gid
--    return res, uid, gid
--end
--
--local function check()
--
--    local is_install_docker = cache.get_global_cache('install_docker')
--    if is_install_docker == nil or not is_install_docker then
--        return
--    end
--
--    local cache_key='docker.service'
--    local old = cache.get_cache(cache_key)
--    if old == nil then
--        local res, uid, gid = is_root('/usr/lib/systemd/system/docker.service')
--        if res > 0 then
--            trigger({Uid=tostring(uid),Gid=tostring(gid)})
--        end
--        cache.set_cache(cache_key, res)
--        return
--    end
--    local currents, uid, gid = is_root('/usr/lib/systemd/system/docker.service')
--    if old ~= currents  then
--        if currents > 0 then
--            trigger({Uid=tostring(uid),Gid=tostring(gid)})
--        end
--        cache.set_cache(cache_key, currents)
--    end
--
--end
--check()
