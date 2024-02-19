--[[
    #![options({
        age: int = 1,
        ageO?: int = 1,
        position = "leader",
        positionO? = "leader",
        nameO?: string,
        lastnameO?
    })]
--]]


function main()
    -- implement global variables
    api.info.log(age)

    return true
end
