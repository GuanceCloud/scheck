valuemonitor = {}

function valuemonitor.check(cache_key, value, content)

	local old = get_cache(cache_key)

	if old == nil then
		set_cache(cache_key, value)
		return
	end
	if old ~= value then
		trigger({Content=tostring(content)})
		set_cache(cache_key, value)
	end
end



return valuemonitor