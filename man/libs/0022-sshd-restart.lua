local system = require("system")
local cache = require("cache")
local function get_sshpid()
    local processes = system.processes()
    for i,v in ipairs(processes) do
        if v["name"] == 'sshd' and v['cmdline'] == '/usr/sbin/sshd -D'  then
            return v["pid"]
        end
    end
    return nil
end

local function check()

    local cache_key="sshd"
    local old = cache.get_cache(cache_key)
    if old == nil then
        local current = get_sshpid()
        if current == nil then
            return
        end
        cache.set_cache(cache_key, current)
        return
    end
    local current = get_sshpid()
    if current == nil then
        return
    end
    if  current ~= old then
        trigger({Pid=tostring(current)})
        cache.set_cache(cache_key, current)
    end

end

check()

