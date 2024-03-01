#![Unsafe]

local spitoTestPath = "/tmp/spito-test"
local fileToRevertPath = spitoTestPath .. "/2fr4738gh5132"

function main()
    -- I create it manyally because I want to test the revert function
    _, err = api.sh.command("mkdir " .. spitoTestPath)
    if err then
    	api.info.error(err)
    	return false
    end
    
    _, err = api.sh.command("echo 'It should be reverted by revert function' > " .. fileToRevertPath)
    if err then
    	api.info.error(err)
    	return false
    end

	return true
end

function revert()
    _, err = api.sh.command("rm " .. fileToRevertPath)
    if err then
    	api.info.error(err)
    	return false
    end
	
	return true
end