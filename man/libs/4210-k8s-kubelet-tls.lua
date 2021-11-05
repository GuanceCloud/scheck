
local cache = require("cache")
local system = require("system")

function check()
    local kube_node = cache.get_global_cache('kube_node')
    if kube_node == nil or not kube_node then
        return
    end

    local tlsfileVal = get_config("tls-cert-file")
    local keyfileVal = get_config("tls-private-key-file")

    if tlsfileVal == "" or keyfileVal == "" then
        local tlsfileCache = cache.get_cache("tls-cert-file")
        local keyfileCache = cache.get_cache("tls-private-key-file")

        if tlsfileCache == nil or keyfileCache == nil then
            trigger()
            cache.set_cache("tls-cert-file",tlsfileVal)
            cache.set_cache("tls-private-key-file",keyfileVal)
        end
    end
end

function get_config(command)
    return system.process_command("kubelet",command)
end

check()
