
local cache = require("cache")
local system = require("system")

function check()
    local kube_node = cache.get_global_cache('apiserver')
    if kube_node == nil or not kube_node then
        return
    end
    local conf = get_config()
    cache_mode = cache.get_cache("anonymous-auth")

    if conf ~= "false" then
        if not cache_mode or cache_mode ~= conf then
            trigger()
            cache.set_cache("anonymous-auth",conf)
        end
    end
end

function get_config()
    return system.process_command("kube-apiserver","anonymous-auth")
end

check()
