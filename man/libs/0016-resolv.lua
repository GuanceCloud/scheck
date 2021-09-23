local filemonitor = require("filemonitor")
local function check()
    filemonitor.check('/etc/resolv.conf')
end

check()