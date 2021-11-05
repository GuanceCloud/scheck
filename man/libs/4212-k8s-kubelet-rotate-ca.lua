local cache = require("cache")
local system = require("system")
local utils = require("utils")
function check()
    local kube_node = cache.get_global_cache('kube_node')
    if kube_node == nil or not kube_node then
        return
    end
    local conf = get_config()
    cache_mode = cache.get_cache("RotateKubeletServerCertificate")
    local val = utils.get_command_value(conf,"RotateKubeletServerCertificate")
    if val  then
        if not cache_mode or cache_mode ~= val then
            trigger()
            cache.set_cache("RotateKubeletServerCertificate",val)
        end
    end
end

function get_config()
    return system.process_command("kubelet","feature-gates")
end

check()
