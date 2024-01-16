function main()
    require_remote("avorty/spito-ruleset", "dbus")
    require_file('./external_file_test.lua')
    return test()
end
