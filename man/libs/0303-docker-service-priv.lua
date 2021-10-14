local sc_file = require("file")
local cache = require("cache")
local filemonitor = require("filemonitor")

local function check(file)
    local is_install_docker = cache.get_global_cache('install_docker')
    if is_install_docker == nil or not is_install_docker then
        return
    end
    if not sc_file.file_exist(file) then
        return
    end
    -- 000 110 100 100   420
    -- --- rw- r-- r--
    filemonitor.priv_limit(file, 420)
end
file = '/usr/lib/systemd/system/docker.service'

check(file)
