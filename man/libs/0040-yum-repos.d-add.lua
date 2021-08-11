local directorymonitor = require("directorymonitor")
local function check()
    directorymonitor.add("/etc/yum.repos.d/")
end
check()