local filemonitor = require("filemonitor")
local function check()
    filemonitor.check('/etc/rc.local')
end

check()
