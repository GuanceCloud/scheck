local filemonitor = require("filemonitor")
local function check()
	filemonitor.check('/etc/ssh/sshd_config')
end
check()