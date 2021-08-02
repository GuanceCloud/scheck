local sysctlmonitor = require("sysctlmonitor")
local function check()
    sysctlmonitor.check('fs.suid_dumpable', 0)
end
check()