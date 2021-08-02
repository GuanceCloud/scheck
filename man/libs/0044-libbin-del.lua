local directorymonitor = require("directorymonitor")

local function check()

directorymonitor.del("/bin")
end
check()