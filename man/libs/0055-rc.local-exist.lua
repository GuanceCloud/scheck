local filemonitor = require("filemonitor")
local function check()
    filemonitor.exist('/etc/rc.local')
end
check()
