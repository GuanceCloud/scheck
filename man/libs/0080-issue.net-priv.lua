local filemonitor = require("filemonitor")
local function check()
    filemonitor.priv_fixedchange('/etc/issue.net', '-rw-r--r--')
end
check()