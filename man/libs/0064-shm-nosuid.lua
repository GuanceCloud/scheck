local mountflagmonitor = require("mountflagmonitor")
mountflagmonitor.check('/dev/shm','nosuid',true)

--mountflagmonitor.check(mountpath, mountflag, value)
--mountpath:需检测的mount 路径名称
--value：yes表示应该设置（未设置或取消设置会trigger），no表示不应该设置（已设置或设置后会trigger）
--mountflag: The options for mounting a file system are generally divided into:
-- async: async mode;
-- sync: Sync mode;
-- atime/noatime: contains directories and files
-- didiRatime/nodiRatime: The access timestamp of the directory
-- auto/noauto: Whether auto mount is supported
-- exec/noexec: Supports running applications on the file system as processes
-- dev/nodev: Whether device files are supported on this file system
-- suid/nosuid: Whether special permissions are supported on this file system
-- Remount: Remount
-- RO: Read only
-- rw: reading and writing
-- user/nouser: Is normal user allowed to mount this device
-- ACL: Enable ACL functionality on this file system
-- Note: More than one option can be used at the same time, separated by commas
--Mount options: defaults: rw, suid, dev, exec, auto, nouser, and async