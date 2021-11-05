local cache = require("cache")
local filemonitor = require("filemonitor")
local function check(files)
    local is_install_apiserver = cache.get_global_cache('kube-apiserver')
    if is_install_apiserver == nil or not is_install_apiserver then
        return
    end
    for i, file in ipairs(files) do
        filemonitor.priv_ownership(file,"etcd")
    end
end

files = {"/var/lib/etcd"}
check(files)
