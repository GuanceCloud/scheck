local directorymonitor = require("directorymonitor")
local dirs={
    '/usr/sbin'
}

for i,v in ipairs(dirs) do
	directorymonitor.priv_change(v)
end

