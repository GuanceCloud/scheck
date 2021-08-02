local kernelmonitor = require("kernelmonitor")
local function check()
    kernelmonitor.module_isinstall('cramfs', false)
end

check()

