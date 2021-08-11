local filemonitor = require("filemonitor")

local function check()
    filemonitor.exist('/etc/hosts')
end


