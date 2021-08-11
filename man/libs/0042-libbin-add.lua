local directorymonitor = require("directorymonitor")

local function check()

directorymonitor.add("/bin")
end
check()