---
--- Generated by Luanalysis
--- Created by 18332.
--- DateTime: 2021/10/25 10:56

local function check()
    local docker = require("container")
    if docker.sc_docker_exist() then
      local cons= docker.sc_docker_containers()
        for i, v in ipairs(cons) do
            id = v.ID
            if id ~= "" then
              run =  docker.sc_docker_runlike(id)
                print(run)
            end
        end
       local images = docker.sc_docker_images()
        print(#images)
    else
        print("docker not exist")
    end
    -- todo 正常的情况下 返回的格式如下： docker run --detach --name nginx-test -p 8080:80 --hostname 5e252952e0f7 --log-driver=journald nginx nginx -g daemon off;
end


check()