local test = require("test")
local system = require("system")
local cache = require("cache")
local mysql = require("mysql")
local function mysql_weak_psw_port(port)
  --for k,v in ipairs(ports) do
     local res, ress = mysql.mysql_weak_psw('localhost', tostring(port))
     if res then
      return ress
     end
  --end

end

local function is_mysqld()
    local processes = system.processes()
    for i,v in ipairs(processes) do
        local res = string.find(v['cmdline'], "mysqld")
        if  res ~= nil  then
            return true
        end
    end
    return false

end



local function check()
    --if not is_mysqld() then
    --    return
    --end
    local is_install_mysql = cache.get_global_cache('install_mysql')
    if is_install_mysql == nil or not is_install_mysql then
        return
    end

    local cache_key="mysql"
    local old = cache.get_cache(cache_key)
    local mysqls = mysql.mysql_ports_list()
    for ii,vv in ipairs(mysqls) do
        local port = ''
        local user = ''
        local pid = ''
        --for i,v in ipairs(mytable) do
        local res = mysql_weak_psw_port(vv['port'])
        if res ~= "" then
            if port ~= '' then port=port.."; " end
            if user ~= '' then user=user.."; " end
            if pid ~= '' then pid=pid.."; " end
            port=port..vv['port']
            user='root'
            pid=pid..vv['pid']
        end
        --end
        --print(pid)
        if port ~= "" then
            trigger({Port=port,User=user,Pid=tostring(pid)})
        end
    end



end

check()
