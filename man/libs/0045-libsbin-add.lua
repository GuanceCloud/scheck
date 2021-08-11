local directorymonitor = require("directorymonitor")

local function check()

directorymonitor.add("/sbin")
end
check()