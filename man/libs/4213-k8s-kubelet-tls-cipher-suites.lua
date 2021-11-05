local cache = require("cache")
local system = require("system")
local k8s = require("container")
function check()
    local kube_node = cache.get_global_cache('kube_node')
    if kube_node == nil or not kube_node then
        return
    end
    local conf = get_config()
    cache_mode = cache.get_cache("tls-cipher-suites")
    local cmd = k8s.sc_tls_cipher_suites(conf)
    if not cmd  then
        if  cache_mode == nil or cache_mode ~= cmd  then
            trigger()
            cache.set_cache("tls-cipher-suites",false)
            return
        end
    end
end

function get_config()
    return system.process_command("kubelet","tls-cipher-suites")
end

check()
