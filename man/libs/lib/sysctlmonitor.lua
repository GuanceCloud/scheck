sysctlmonitor = {}

-- Get specific kernel parameters
function get_value(list, param)
    for ii,vv in pairs(list) do
        if ii == param then
            return vv
        end
    end
end

-- This is the method to check whether the kernel parameters are up to standard. Param is the kernel parameter name and switch is the final value
function sysctlmonitor.check(param, switch)
    local cache_key = param
    local current = sysctl(param)[param]
    switch = tostring(switch)

    local old=get_cache(cache_key)

    if old == nil   then
        if current ~= switch then
            trigger({Content=string.format("%s = %s ", param, current)})
        end
        set_cache(cache_key, current)
        return
    end

    if  current ~= old and current ~= switch then
        trigger({Content=string.format("%s = %s ", param, current)})
        set_cache(cache_key, current)
    end

end

function sysctlmonitor.check_many(params, switch)
    local sysctlkey = 'sysctl'
    local sysctllist = "has"
    local oldsysctl = get_cache(sysctlkey)

    if oldsysctl == nil then
        local content = ''
        for key,value in ipairs(params) do
            local current = sysctl(value)[value]
            if current ==nil then current='' end

            if current ~= switch then
                if content ~= '' then content=content.."; " end
                content = content..string.format("%s = %s ", value, current)
            end
            set_cache(value, current)
        end
        if content ~= ''
        then
            trigger({Content=content})
        end
        set_cache(sysctlkey, sysctllist)
        return
    end
    local content=''
    for key,value in ipairs(params) do
        local current = sysctl(value)[value]
        local old = get_cache(value)
        if current ==nil then current='' end
        if old ~= current  and current ~= switch then
            if content ~= '' then content=content.."; " end
            content = content..string.format("%s = %s ", value, current)
        end
        set_cache(value, current)
    end
    if content ~= ''
    then
        trigger({Content=content})
    end
end

return sysctlmonitor