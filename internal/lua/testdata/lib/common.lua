common = {}
local system = require("system")
local sc_file = require("file")
function Field_filtering(data)
    local tabT={}
    local content = ''
    for key,v in pairs(data) do
        if tabT[data[key]] == nil then
            tabT[data[key]] = 1
        else
            tabT[data[key]] = tabT[data[key]] + 1
        end
    end
    for i,v in pairs(tabT) do
        if content ~= '' then content=content.."; " end
        if i == 1 then
            content = content ..string.format("CREATE %d", v)
        elseif i == 2 then
            content = content ..string.format("WRITE %d", v)
        elseif i == 4 then
            content = content ..string.format("REMOVE %d", v)
        elseif i == 8 then
            content = content ..string.format("RENAME %d", v)
        elseif i == 16 then
            content = content ..string.format("CHMOD %d", v)
        end
    end
    return content
end

function common.watcher(path, enum)
    if not sc_file.file_exist(path)  then
        return
    end
    local ch3 = channel.make()
    sc_file.sc_path_watch(path, ch3)
    local ticker = channel.make()
    system.sc_ticker(ticker)
    local exit = false
    local data = {}
    while not exit do
        channel.select(
                {"|<-", ch3, function(ok, v)
                    if ok then
                        if data[v['path']] == nil then
                            local path = {}
                            path[#path+1] = v['status']
                            data[v['path']] = path
                        else
                            data[v['path']][#data[v['path']]+1] = v['status']
                        end
                    end
                end},
                {"|<-", ticker, function(ok, v)
                    if ok then
                        if next(data) ~= nil then
                            for key, value in pairs(data) do
                                local flag = false
                                local isQuit = false
                                for k, v in pairs(value) do
                                    for kk,vv in pairs(enum) do
                                        if v == vv then
                                            flag = true
                                        end
                                    end
                                    if v == 4 or v == 8  then
                                        if key == path then
                                            isQuit = true
                                        end
                                    end
                                end
                                if flag then
                                    trigger({Content=string.format("%s change, %s:%s ", path, key, Field_filtering(value))})
                                    data = {}
                                end
                                if isQuit then
                                    exit = true
                                end
                            end
                        end
                    end
                end
                }
        )
    end
end

return common
