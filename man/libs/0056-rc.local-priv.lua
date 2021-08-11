local filemonitor = require("filemonitor")
local files={
    '/etc/rc.d/rc.local',
    '/etc/rc.local'
}

for i,v in ipairs(files) do
	filemonitor.priv_change(v)
end

