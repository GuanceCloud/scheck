local sc_file = require("file")
local cache = require("cache")
local system = require("system")
local  function check()
    local key = "authorized_keys"
    local userlist = system.users()
    local oldvalue = cache.get_cache(key)
    if oldvalue == nil then
        for ii,value in ipairs(userlist) do
            if value['uid'] > 1000 or value["uid"] == 0 then
                local authorized_keys_path = string.format("%s/%s", value["directory"], ".ssh/authorized_keys")
                if sc_file.file_exist(authorized_keys_path) then
                    cache.set_cache(authorized_keys_path, sc_file.file_hash(authorized_keys_path))
                end
            end
        end
        cache.set_cache(key, userlist)
        return
    end
    local content=''
    local usename=''
    for ii,value in ipairs(oldvalue) do
        if value['uid'] > 1000 or value["uid"] == 0 then
            local authorized_keys_path = string.format("%s/%s", value["directory"], ".ssh/authorized_keys")
            if sc_file.file_exist(authorized_keys_path) then
                local filehash = sc_file.file_hash(authorized_keys_path)
                local old = cache.get_cache(authorized_keys_path)
                if filehash ~= old then
                    if content ~= '' then content=content.."; " end
                    if usename ~= '' then usename=usename.."; " end
                    content = content..string.format("%s", authorized_keys_path)
                    usename = usename..string.format("%s", value['username'])
                    cache.set_cache(authorized_keys_path, filehash)
                end
            end
        end
    end

    if content ~= ''
    then
        trigger({File=content,User=usename})
    end
    if #userlist ~= oldvalue then
      cache.set_cache(key, userlist)
    end

end


check()