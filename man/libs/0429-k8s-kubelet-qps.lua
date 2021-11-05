
local cache = require("cache")
local system = require("system")

function check()
    local kube_node = cache.get_global_cache('kube_node')
    if kube_node == nil or not kube_node then
        return
    end
    local conf = get_config()
   local cache_mode = cache.get_cache("event-qps")

    if conf == "" then
        -- 默认为5 可以直接退出
        return
    end
   local count = tonumber(conf)
    if count < 5 then
        if not cache_mode or cache_mode < count then
            trigger({Content=count})
            cache.set_cache("event-qps",count)
        end
    end
end

function get_config()
    return system.process_command("kubelet","event-qps")
end

check()
