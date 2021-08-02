local kernelmonitor = require("kernelmonitor")
local function check()
    kernelmonitor.module_isinstall('squashfs', false)
end

check()

