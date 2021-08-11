local rpmmonitor = require("rpmmonitor")


--value：true表示应该安装（卸载后会trigger），false表示不应该安装（安装后会trigger）
rpmmonitor.check('xinetd',false)
