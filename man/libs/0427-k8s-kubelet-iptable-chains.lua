
local cache = require("cache")
local system = require("system")

function check()
    local kube_node = cache.get_global_cache('kube_node')
    if kube_node == nil or not kube_node then
        return
    end
    local conf = get_config()
    cache_mode = cache.get_cache("make-iptables-util-chains")
    if conf == "" then
        return
    end
    if conf ~= "true" then
        if not cache_mode or cache_mode ~= conf then
            trigger({Content=conf})
            cache.set_cache("make-iptables-util-chains",conf)
        end
    end
end

function get_config()
    return system.process_command("kubelet","make-iptables-util-chains")
end

check()
