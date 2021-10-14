directorymonitor = {}
require("common")
local system = require("system")
local cache = require("cache")
local sc_file = require("file")
-- This method is to detect whether in the directory file md5 has changed,and the parameter is the directory path
function directorymonitor.change(dir)
    local exit = false
    while not exit do
        if sc_file.file_exist(dir)  then
            common.watcher(dir, {2, 8}, common.file_trigger)
        end
        system.sc_sleep(5)
    end
end

-- This method is to detect whether in the directory file md5 has changed,and the parameter is the directory path
function directorymonitor.add(dir)
    local exit = false
    while not exit do
        if sc_file.file_exist(dir)  then
            common.watcher(dir, {1}, common.file_trigger)
        end
        system.sc_sleep(5)
    end
end

function directorymonitor.del(dir)
    local exit = false
    while not exit do
        if sc_file.file_exist(dir)  then
            common.watcher(dir, {4}, common.file_trigger)
        end
        system.sc_sleep(5)
    end
end


function directorymonitor.priv_change(dir)
    local exit = false
    while not exit do
        if sc_file.file_exist(dir)  then
            common.watcher(dir, {16}, common.file_trigger)
        end
        system.sc_sleep(5)
    end
end

return directorymonitor
