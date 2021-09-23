local filemonitor = require("filemonitor")
local function check()
    filemonitor.priv_fixedchange('/boot/grub2/grub.cfg', '-rw-------')
end
check()