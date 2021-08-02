local filemonitor = require("filemonitor")
local function check()
    filemonitor.check('/etc/fstab')
end

check()