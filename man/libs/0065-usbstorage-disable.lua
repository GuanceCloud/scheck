local kernelmonitor = require("kernelmonitor")
local function check()
    kernelmonitor.module_isinstall('usb_storage', false)
end
check()
