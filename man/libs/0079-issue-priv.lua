local filemonitor = require("filemonitor")
local function check()
    filemonitor.priv_fixedchange('/etc/issue', '-rw-r--r--')
end
check()