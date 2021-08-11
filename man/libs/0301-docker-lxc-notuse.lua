
local function  is_use_lxc()
    local processes = processes()
    for i,v in ipairs(processes) do
        local res1 = string.find(v['cmdline'], "docker")
        local res2 = string.find(v['cmdline'], "lxc")
        if  res1 ~= nil and res2 ~= nil then
            return true
        end
    end
    return nil
end

local function check()
    local is_install_docker = get_global_cache('install_docker')
    if is_install_docker == nil or not is_install_docker then
        return
    end

    local cache_key = "lxc"
    --local value = kernel_info()['version']
    --local current = is_use_lxc()
    local old = get_cache(cache_key)

    if old == nil then
        local current = is_use_lxc()
        if current ~= nil then
            trigger()
        end
        set_cache(cache_key, current)
        return
    end

    local current = is_use_lxc()
    if old ~= current then
        trigger()
        set_cache(cache_key, current)
    end
end
check()
