install_mysql = false
install_docker = false
kube_node = false
kube_apiserver = false
etcd = false
scheduler = false
local system = require("system")
local cache = require("cache")
local container =  require("container")
local function check()
    local processes = system.processes()
    for i,v in ipairs(processes) do
        if  string.find(v['cmdline'], "mysqld") ~= nil then
            install_mysql = true
        end
        if  string.find(v['cmdline'], "docker") ~= nil then
            install_docker = true
        end
        if  string.find(v['cmdline'], "kubelet") ~= nil then
          local checkV= container.sc_kubectl_checkVersion()
            if checkV then
                kube_node = true
            end
        end
        if  string.find(v['cmdline'], "kube-apiserver") ~= nil then
            kube_apiserver = true
        end
        if  string.find(v['cmdline'], "kube-scheduler") ~= nil then
            scheduler = true
        end
        if  string.find(v['cmdline'], "etcd") ~= nil then
            etcd = true
        end
    end
    cache.set_global_cache('install_mysql', install_mysql)
    cache.set_global_cache('install_docker', install_docker)
    cache.set_global_cache('kube_node', kube_node)
    cache.set_global_cache('apiserver', kube_apiserver)
    cache.set_global_cache('scheduler', scheduler)
    cache.set_global_cache('etcd', etcd)
end
check()
