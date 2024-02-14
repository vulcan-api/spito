--[[
    #![options({
        age: int = 1,
        name: string,
        lastname,
        position = "leader",
    })]
--]]


function main()
    -- implement global variables
    api.info.log(age)

    return true
end
