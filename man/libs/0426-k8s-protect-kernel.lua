
local cache = require("cache")
local system = require("system")

function check()
    local kube_node = cache.get_global_cache('kube_node')
    if kube_node == nil or not kube_node then
        return
    end
    local conf = get_config()
    cache_mode = cache.get_cache("protect-kernel-defaults")

    if conf ~= "true" then
        if not cache_mode or cache_mode ~= conf then
            trigger()
            cache.set_cache("protect-kernel-defaults",conf)
        end
    end
end

function get_config()
    return system.process_command("kubelet","protect-kernel-defaults")
end

check()
