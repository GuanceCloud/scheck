local cache = require("cache")
local filemonitor = require("filemonitor")
local function check(files)
    local is_install_apiserver = cache.get_global_cache('apiserver')
    if is_install_apiserver == nil or not is_install_apiserver then
        return
    end
    for i, file in ipairs(files) do
        filemonitor.priv_root_ownership(file)
    end
   
end

files = {"/etc/kubernetes/manifests/kube-apiserver.yaml","/etc/kubernetes/manifests/kube-controller-manager.yaml","/etc/kubernetes/manifests/kube-scheduler.yaml"}
check(files)
