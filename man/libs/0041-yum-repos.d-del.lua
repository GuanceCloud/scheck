local directorymonitor = require("directorymonitor")
local function check()
    directorymonitor.del("/etc/yum.repos.d/")
end
check()