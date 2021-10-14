kernelmonitor = {}
local system = require("system")
local cache = require("cache")
function judge_modules_installed(modules, module_name)
    local exist = false
    for key, value in ipairs(modules) do
        if value["name"] == module_name then
            exist = true
            break
        end
    end
    return exist

end

-- Whether the kernel module needs to be installed, module_ Name is the module name，isInstalled is bool false true
function kernelmonitor.module_isinstall(module_name, isInstalled)
    local cache_key = module_name
    local old = cache.get_cache(cache_key)

    local status = ''
    if isInstalled then status = "uninstalled" else status = "installed"  end

    if old == nil then
        local currents = judge_modules_installed(system.kernel_modules(), module_name)
        if currents ~= isInstalled then
           -- print(string.format("kernel %s is %s ", module_name, status ))
            trigger({Content=string.format("kernel %s is %s ", module_name, status )})
        end
        cache.set_cache(cache_key, currents)
        local old = cache.get_cache(cache_key)
        return
    end
    local currents = judge_modules_installed(system.kernel_modules(), module_name)
    if old ~= currents then
        if currents ~= isInstalled then
            trigger({Content=string.format("kernel %s is %s ", module_name, status )})
        end
        cache.set_cache(cache_key, currents)
    end
end



return kernelmonitor
