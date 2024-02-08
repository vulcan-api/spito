#![unsafe]

function main()
    res = api.sh.command("echo \"Test\"")
    
    return trim(res) == "Test"
end

function trim(s)
    return (s:gsub("^%s*(.-)%s*$", "%1"))
end
