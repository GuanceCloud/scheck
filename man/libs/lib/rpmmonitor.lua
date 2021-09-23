rpmmonitor = {}
--For example, make sure sudo(rpm) is installed
function rpm_testing(pck_name)
    local data = rpm_query(pck_name)
    if data == nil or data =='' then
        return "false"
    else
        return "true"
    end
end

function rpmmonitor.check(pck_name, switch)
    switch = tostring(switch)
    if switch ~= "true" or switch ~= "false" then
        return
    end

    local cache_key = pck_name
    local old=get_cache(cache_key)

    if old == nil   then
        local currents = rpm_testing(pck_name)
        if currents ~= switch then
            trigger({Content=pck_name})
        end
        set_cache(cache_key, current)
        return
    end
    local currents = rpm_testing(pck_name)
    if  current ~= old and currents ~= switch then
        trigger({Content=pck_name})
        set_cache(cache_key, currents)
    end
end



--function rpmmonitor.check(pck, value)
--    local rpm_pck = rpm_query(pck)
--
-- --Determine whether the installation package is currently installed
-- --It returns true if installed (with a return value) and false if not installed (with no return)
--    if rpm_pck ~= '' then
--    --('package '..pck..' is not installed')
--        set_cache('rpm_installed',true)
--    else
--        set_cache('rpm_installed',false)
--    end
--
--	local current = get_cache('rpm_installed')
--    local old_trigger = get_cache('rpm_trigger')
-- --     if get_cache(rpm_trigger) == nil then
-- --         set_cache(rpm_trigger,true)
-- --     end
--
-- --Determine whether the current installation value of the package is consistent with the base value
--    if current == value then
--        --set_cache(rpm_installed,current)
--        set_cache('rpm_trigger','normal')
--        return
--    --Determine the last alarm status to prevent repeated alarms
--    elseif old_trigger == 'normal' or old_trigger == nil then
--    --If the current value is not consistent with the base value, you need to trigger.
--            if current == true
--            then
--                trigger({Content='RPMs：'..pck})
--                --trigger({Content='Found RPMs '..pck..' with security risks,Suggest Uninstalling'})
--                --set_cache(rpm_installed,current)
--                set_cache('rpm_trigger','error')
--            --If the current value is not consistent with the base value, trigger is required.
--            elseif current == false
--            then
--                trigger({Content='RPMs：'..pck})
--                --trigger({Content='Found RPMs：'..pck..' not installed，Suggest Installing'})
--                --set_cache(rpm_installed,current)
--                set_cache('rpm_trigger','error')
--            end
--    end
--end
return rpmmonitor
