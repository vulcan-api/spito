-- This is function is executed by default
function main()
    -- api is the only way to interact with machine
    api.info.Log("Hello world!")
    
    -- True means that the rule pass
    return true
end