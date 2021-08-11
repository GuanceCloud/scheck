local filemonitor = require("filemonitor")

local function check(file)
    local is_install_docker = get_global_cache('install_docker')
    if is_install_docker == nil or not is_install_docker then
        return
    end
    if not file_exist(file) then
        return
    end
    -- 000 110 100 100   420
    -- --- rw- r-- r--
    filemonitor.priv_limit(file, 420)
end
file = '/usr/lib/systemd/system/docker-registry.service'

check(file)

