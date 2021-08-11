local directorymonitor = require("directorymonitor")

local function check()

directorymonitor.del("/usr/sbin")
end
check()