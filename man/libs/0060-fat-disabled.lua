local kernelmonitor = require("kernelmonitor")
local function check()
    kernelmonitor.module_isinstall('vfat', false)
end

check()

