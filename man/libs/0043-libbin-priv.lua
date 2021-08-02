local directorymonitor = require("directorymonitor")
local dirs={
    '/bin'
}
for i,v in ipairs(dirs) do
	directorymonitor.priv_change(v)
end

