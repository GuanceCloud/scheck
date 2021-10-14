local filemonitor = require("filemonitor")
local cache = require("cache")
local function check(file)
    local is_install_docker = cache.get_global_cache('install_docker')
    if is_install_docker == nil or not is_install_docker then
        return
    end
    filemonitor.priv_root_ownership(file)
end

local file = '/usr/lib/systemd/system/docker-registry.service'
check(file)
