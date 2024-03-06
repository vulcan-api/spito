#![Unsafe]

local spitoTestPath = "/tmp/spito-test"
local fileToRevertPath = spitoTestPath .. "/2fr4738gh5132"

function main()
    -- I create it manually because I want to test the revert function
    local _, err = api.sh.command("mkdir -p " .. spitoTestPath)
    if err then
        api.info.log("____________________________________________________________--")
        api.info.error(err)
        return false
    end

    local _, err = api.sh.command("echo 'It should be reverted by revert function' > " .. fileToRevertPath)
    if err then
        api.info.error(err)
        return false
    end

    return true
end

function revert()
    local _, err = api.sh.command("rm " .. fileToRevertPath)
    if err then
        api.info.error(err)
        return false
    end

    return true
end
