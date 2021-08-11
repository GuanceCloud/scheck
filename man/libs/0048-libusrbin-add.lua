local directorymonitor = require("directorymonitor")

local function check()

directorymonitor.add("/usr/bin")
end
check()