local filemonitor = require("filemonitor")
local function check()
    filemonitor.priv_fixedchange('/etc/motd', '-rw-r--r--')
end
check()