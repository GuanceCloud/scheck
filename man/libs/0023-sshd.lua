local filemonitor = require("filemonitor")
local function check()
	filemonitor.check('/etc/ssh/sshd_config')
end
check()
--local files={
--    '/etc/ssh/sshd_config',
--
--}
--local function check(file)
--	local cache_key=file
--	local hashval = file_hash(file)
--
--	local old = get_cache(cache_key)
--	if not old then
--		set_cache(cache_key, hashval)
--		return
--	end
--
--	if old ~= hashval then
--		trigger({File=file})
--		set_cache(cache_key, hashval)
--	end
--end
--
--for i,v in ipairs(files) do
--	check(v)
--end
--

