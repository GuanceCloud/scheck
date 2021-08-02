function grep_sudofile(file)
    if file_exist(file) then
        local data = grep("-Ei","^\\s*Defaults\\s+([^#]\\S+,\\s*)?use_pty\\b", file)
        if data ~= '' then
            return true
        end
    end
    return false
end

function use_pty()
   local flag = false
   if file_exist("/etc/sudoers") then
       if grep_sudofile("/etc/sudoers") then
           flag = true
       end
   end
   if file_exist("/etc/sudoers.d/") then
        local filelist = ls("/etc/sudoers.d/")
        for ii,value in ipairs(filelist) do
            if grep_sudofile(value) then
                flag = true
                break
            end
        end
   end
   return flag
end

local function check()
    local cache_key = "use_pty"
    local current = use_pty()
    print(current)
    local old = get_cache(cache_key)

    if old == nil then
        if not current
        then
            trigger()
        end
        set_cache(cache_key, current)
        return
    end

    if old ~= current then
        if not current
        then
            trigger()
        end
        set_cache(cache_key, current)
    end

end
check()