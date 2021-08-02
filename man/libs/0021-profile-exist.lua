local filemonitor = require("filemonitor")
local function check()
    filemonitor.exist('/etc/profile')
end
check()