local directorymonitor = require("directorymonitor")
local function check()
    directorymonitor.change("/etc/yum.repos.d")
end
check()
