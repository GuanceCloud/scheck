local function check()
    local cache_key='check_ports'
    local old=get_cache(cache_key)
    if not old then
        local ports = listening_ports()
        set_cache(cache_key, ports)
        return
    end

    local content=''
    local ports = listening_ports()
    for i,v in ipairs(ports) do
        if v["family"] == 'AF_INET' then
            local exist=false
            for ii,vv in ipairs(old) do
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
        set_cache(cache_key, ports)
    end

end

check()
