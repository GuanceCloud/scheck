local filemonitor = require("filemonitor")
local function check()
    filemonitor.check('/etc/sudoers')
end

check()
