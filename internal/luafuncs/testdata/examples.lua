function GetArray()
  local files = put_to_table()
   for ii,vv in ipairs(files) do
      if vv["x100"] ~=""
      then
         vv["x100"] =""
      end
   end
   --print(#files)
end

GetArray()