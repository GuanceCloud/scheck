local function get_tunnel()
    local processes = processes()
    local count = 0
    local pid = ''
     for i,v in ipairs(processes) do
        if v["name"] == 'sshd' and v['cmdline'] == 'sshd: root@notty'  then
            count = count + 1
            if pid ~= '' then pid=pid..";" end
            pid = pid..v["pid"]
        end
     end
    return count,pid

end

local function check()

    local cache_key="sshd_tunnel"
    local currents,pid = get_tunnel()
    local old = get_cache(cache_key)
    if old == nil then
        if currents > 0 then
            trigger({Count=tostring(currents),Pid=pid})
        end
        set_cache(cache_key, currents)
        return
    end

   	if  currents ~= old then
        if currents > 0 then
            trigger({Count=tostring(currents),Pid=pid})
        end
   		set_cache(cache_key, currents)
   	end

end



check()
