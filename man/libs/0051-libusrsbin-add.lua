local directorymonitor = require("directorymonitor")

local function check()

directorymonitor.add("/usr/sbin")
end
check()