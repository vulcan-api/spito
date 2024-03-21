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
        family:list={1;2;3},
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
        { name = "family", subOptions = { { name = "0", value = "1" }, { name = "1", value = "2" }, { name = "2", value = "3" } } },
        { name = "dog", subOptions = { { name = "hairType" }, { name = "age", value = 5 } } }
    }
    return checkArray(testArray, OPTIONS)
end

function checkArray(array, rootVariable)
    for i = 1, #array do
        local case = array[i]
        if case.subOptions ~= nil then
            local success = checkArray(case.subOptions, rootVariable[case.name])
            if success ~= true then
                return false
            end
        else
            if rootVariable[case.name] ~= case.value then
                api.info.error("variable named", case.name, "doesn't match wanted value:", _O[case.name], "vs", case.value)
                return false
            end
        end
    end
    return true
end
