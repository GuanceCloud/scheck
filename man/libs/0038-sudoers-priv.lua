local filemonitor = require("filemonitor")
local files={
    '/etc/sudoers'
}

for i,v in ipairs(files) do
	filemonitor.priv_change(v)
end