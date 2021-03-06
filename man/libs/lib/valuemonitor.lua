valuemonitor = {}
local cache = require("cache")
function valuemonitor.check(cache_key, value, content)

	local old = cache.get_cache(cache_key)

	if old == nil then
		cache.set_cache(cache_key, value)
		return
	end
	if old ~= value then
		trigger({Content=tostring(content)})
		cache.set_cache(cache_key, value)
	end
end



return valuemonitor