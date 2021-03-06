filemonitor = {}
local common = require("common")
local system = require("system")
local cache = require("cache")
local sc_file = require("file")
-- This method is to detect whether the file mode has changed,and the parameter is the file path
function filemonitor.check(file)
	local exit = false
	while not exit do
		if sc_file.file_exist(file)  then
			common.watcher(file, {2, 4, 8, 16}, common.file_trigger)
		end
		system.sc_sleep(5)
	end
end
-- This method is to detect whether the file exists,and the parameter is the file path
function filemonitor.exist(file)
    local cache_key = file

    local old = cache.get_cache(cache_key)

    if old == nil   then
		local current = sc_file.file_exist(file)
		if current then
            cache.set_cache(cache_key, current)
            return
        else
            trigger({Content=file})
            cache.set_cache(cache_key, current)
            return
        end
    end
	local current = sc_file.file_exist(file)
   	if  current ~= old and not current then
   		trigger({Content=file})
   		cache.set_cache(cache_key, current)
   	end
end

-- This method is to detect whether the file permission is changed, and the parameter is the file path
function filemonitor.priv_change(file)
	local cache_key = file
	local old = cache.get_cache(cache_key)

	if old == nil then
		local current = sc_file.file_info(file)['mode']
		cache.set_cache(cache_key, current)
		return
	end
	local current = sc_file.file_info(file)['mode']
	if old ~= current then
		trigger({Content=string.format("%s mode %s change %s", file, old, current )})
		cache.set_cache(cache_key, current)
	end

end

function judge_mode(mod, new_mode)
	if mod == new_mode then
		return true
	else
		return false
	end
end


-- This method is to detect whether the file permission is changed, file parameter input  file path example /etc/hosts; mode input file mode example -rw-r--r--

function filemonitor.priv_fixedchange(file, mode)
	local cache_key = file
	local old = cache.get_cache(cache_key)

	if old == nil then
		local current = sc_file.file_info(file)['mode']

		if not judge_mode(mode, current)
		then
			trigger({Content=string.format("%s mode is not %s ; not is %s", file, mode, current)})
		end
		cache.set_cache(cache_key, current)
		return
	end
	local current = sc_file.file_info(file)['mode']
	if old ~= current then
		if not judge_mode(mode, current)
		then
			trigger({Content=string.format("%s mode is not %s ; change %s", file, mode, current)})
		end
		cache.set_cache(cache_key, current)
	end

end


function filemonitor.priv_limit(file, mode_bits)
	local cache_key = file
	if not sc_file.file_exist(file) then
		return
	end
	local old = cache.get_cache(cache_key)

	if old == nil then
		local current_key = sc_file.file_info(file)
		local mode = current_key['mode']
		local current = current_key['perm']
		if current < mode_bits
		then
			trigger({Priv=mode})
		end
		cache.set_cache(cache_key, current)
		return
	end

	local current_key = sc_file.file_info(file)
	local mode = current_key['mode']
	local current = current_key['perm']
	if old ~= current then
		if current < mode_bits
		then
			trigger({Priv=mode})
		end
		cache.set_cache(cache_key, current)
	end

end

local function is_root_ownership(file)
	local mode = sc_file.file_info(file)
	local uid = mode['uid']
	local gid = mode['gid']
	local  res = uid + gid
	return res, uid, gid
end

function filemonitor.priv_root_ownership(file)

	if not sc_file.file_exist(file) then
		return
	end
	local cache_key='root_ownership'
	local old = cache.get_cache(cache_key)
	if old == nil then
		local res, uid, gid = is_root_ownership(file)
		if res > 0 then
			trigger({Uid=tostring(uid),Gid=tostring(gid)})
		end
		cache.set_cache(cache_key, res)
		return
	end
	local currents, uid, gid = is_root_ownership(file)
	if old ~= currents  then
		if currents > 0 then
			trigger({Uid=tostring(uid),Gid=tostring(gid)})
		end
		cache.set_cache(cache_key, currents)
	end
end


return filemonitor
