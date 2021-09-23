local kernelmonitor = require("kernelmonitor")
local function check()
    kernelmonitor.module_isinstall('udf', false)
end

check()

