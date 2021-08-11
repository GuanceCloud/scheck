local filemonitor = require("filemonitor")
local function check()
    filemonitor.exist('/etc/fstab')
end
check()