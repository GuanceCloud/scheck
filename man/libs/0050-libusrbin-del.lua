local directorymonitor = require("directorymonitor")

local function check()

directorymonitor.del("/usr/bin")
end
check()