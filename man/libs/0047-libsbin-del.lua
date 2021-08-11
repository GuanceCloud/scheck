local directorymonitor = require("directorymonitor")

local function check()

directorymonitor.del("/sbin")
end
check()