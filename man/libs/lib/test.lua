test = {}
function list_table(data)

    if type(data) == 'table' then
        print('---------------table------------------------')
        for k,v in pairs(data) do
            if isArrayTable(data) then
                list_table(v)
            else
                print(k,v)
            end
        end
        print('-----------------table----------------------' )
    else
        print(string.format('%s is %s', data ,type(data)))
    end
end



function isArrayTable(t)
    if type(t) ~= "table" then
        return false
    end

    local n = #t
    for i,v in pairs(t) do
        if type(i) ~= "number" then
            return false
        end

        if i > n then
            return false
        end
    end

    return true
end




function test.print(data)
    list_table(data)
end



return test