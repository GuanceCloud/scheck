local cache = require("cache")
local function chsize(char)
    if not char then
        --print("not char")
        return 0
    elseif char > 240 then
        return 4
    elseif char > 225 then
        return 3
    elseif char > 192 then
        return 2
    else
        return 1
    end
end

--Calculate the number of UTF8 string characters, all characters are calculated as one character
--For example, utf8len ("1 Hello") = > 3
function utf8len(str)
    local len = 0
    local currentIndex = 1
    while currentIndex <= #str do
        local char = string.byte(str, currentIndex)
        currentIndex = currentIndex + chsize(char)
        len = len +1
    end
    return len
end

--Intercepting UTF8 string
--STR: string to intercept
--Startchar: start character subscript, starting from 1
--Numchars: length of characters to intercept
function utf8sub(str, startChar, numChars)
    local startIndex = 1
    while startChar > 1 do
        local char = string.byte(str, startIndex)
        startIndex = startIndex + chsize(char)
        startChar = startChar - 1
    end

    local currentIndex = startIndex

    while numChars > 0 and currentIndex <= #str do
        local char = string.byte(str, currentIndex)
        currentIndex = currentIndex + chsize(char)
        numChars = numChars -1
    end
    return str:sub(startIndex, currentIndex - 1)
end


local system = require("system")
local function check()
    local is_install_docker = cache.get_global_cache('install_docker')
    if is_install_docker == nil or is_install_docker then
        return
    end

    local cache_key = "docker_kernel_version"
    local current = tonumber(utf8sub( system.kernel_info()['version'], 1, 4))
    local old = cache.get_cache(cache_key)

    if old == nil then
        if current < 3.1 then
            trigger({Version=current})
        end
        cache.set_cache(cache_key, current)
        return
    end

    if old ~= current then
        trigger({Version=current})
        cache.set_cache(cache_key, current)
    end
end
check()
