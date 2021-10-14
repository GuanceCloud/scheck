local net = require("net")
local cache = require("cache")
local function check()
    local cache_key='check_ports'
    local old=cache.get_cache(cache_key)
    if not old then
        local ports = net.listening_ports()
        cache.set_cache(cache_key, ports)
        return
    end

    local content=''
    local ports = net.listening_ports()
    for i,v in ipairs(old) do
        if v["family"] == 'AF_INET' then
            local exist=false
            for ii,vv in ipairs(ports) do
                if vv["port"] == v["port"] and vv["address"] == v["address"]  then
                    exist = true
                    break
                end
            end
            if not exist then
                if content ~= '' then content=content.."; " end
                content = content..string.format("%d(%s) %s/%s", v["port"], v["protocol"], v["pid"],v["process_name"])
            end
        end
    end

    if content ~= '' then
        trigger({Content=content})
        cache.set_cache(cache_key, ports)
    end

end

check()

