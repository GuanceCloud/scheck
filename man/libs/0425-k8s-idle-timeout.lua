
local cache = require("cache")
local system = require("system")

function check()
    local kube_node = cache.get_global_cache('kube_node')
    if kube_node == nil or not kube_node then
        return
    end
    local conf = get_configfile()
    cache_mode = cache.get_cache("streaming-connection-idle-timeout")
    if conf == "0" then
        if not cache_mode then
            trigger({Content=conf})
            cache.set_cache("streaming-connection-idle-timeout",conf)
        end
    end
end

function get_configfile()
    return system.process_command("kubelet","streaming-connection-idle-timeout")
end

check()
