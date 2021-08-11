local filemonitor = require("filemonitor")
local function check(file)
    local is_install_docker = get_global_cache('install_docker')
    if is_install_docker == nil or not is_install_docker then
        return
    end
    filemonitor.priv_root_ownership(file)
end

file = '/etc/sysconfig/docker'
check(file)

