--[[
    #![options({
        ageO?: int = 1,
        position = "leader",
        nameO?: string,
        lastnameO?,
        dog = {
            hairType?,
            age: int = 5
        },
        gender:{male;female}=male,
        positionO? = "leader"
    })]
--]]


function main()
    local testArray = {
        { name = "ageO", value = 1 },
        { name = "position", value = "leader" },
        { name = "nameO" },
        { name = "lastnameO" },
        { name = "positionO", value = "leader" },
        { name = "gender", value = "male" },
        { name = "dog", subOptions = { { name = "hairType", }, { name = "age", value = 5 } } }
    }
    return checkArray(testArray)

end

function checkArray(array)
    for i = 1, #array do
        local case = array[i]
        if case.subOptions ~= nil then
            checkArray(array[i])
        else
            if _O[case.name] ~= case.value then
                api.info.error("variable named", case.name, "doesn't match wanted value:", _O[case.name], "vs", case.value)
                return false
            end
        end
    end
    return true
end