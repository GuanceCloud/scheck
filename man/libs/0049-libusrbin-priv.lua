local directorymonitor = require("directorymonitor")
local dirs={
    '/usr/bin'
}

for i,v in ipairs(dirs) do
	directorymonitor.priv_change(v)
end

