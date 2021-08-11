mountflagmonitor = {}
--For example, make sure sudo(rpm) is installed
function mountflagmonitor.check(mountpath, mountflag, value)
    --local mountpath = '/dev/shm'
    local mountinfo = mounts()
    local mountflag2 = ','..mountflag..','
--Determine whether the installation package is currently installed
--It returns true if installed (with a return value) and false if not installed (with no return)
    for i,v in ipairs(mountinfo) do
        if v['path'] == mountpath then
            flag = ','..v['flags']..','
            if string.match(flag,mountflag2) == nil then
               set_cache('flagexist',false)
            else
               set_cache('flagexist',true)
            end
            --print(string.match(v['flags'],flag))
            --print(i..v['path']..' '..v['device']..' '..v['type']..' '..v['flags'])
            --print(flagexist)
        end
    end

	local current = get_cache('flagexist')
    local old_trigger = get_cache('mountflag_trigger')


--Determine whether the current installation value of the package is consistent with the base value
    if current == value then
        --set_cache(rpm_installed,current)
        set_cache('mountflag_trigger','normal')
    --Determine the last alarm status to prevent repeated alarms
    elseif old_trigger == 'normal' or old_trigger == nil then
    --If the current value is not consistent with the base value, you need to trigger.
            if current == true
            then
                trigger({Content='mountpath：'..mountpath})
                --trigger({Content='Found RPMs '..pck..' with security risks,Suggest Uninstalling'})
                --set_cache(rpm_installed,current)
                set_cache('mountflag_trigger','error')
            --If the current value is not consistent with the base value, trigger is required.
            elseif current == false
            then
                trigger({Content='mountpath：'..mountpath})
                --trigger({Content='Found RPMs：'..pck..' not installed，Suggest Installing'})
                --set_cache(rpm_installed,current)
                set_cache('mountflag_trigger','error')
            elseif current == nil
            then
                trigger({Content='mountpath：'..mountpath..' not exist'})
                set_cache('mountflag_trigger','error')
            end
    end
end
return mountflagmonitor