local sysctlmonitor = require("sysctlmonitor")
local function check()
    sysctlmonitor.check('kernel.randomize_va_space', 2)
end
check()