function main()
    require_remote("https://github.com/avorty/spito-ruleset", "dbus")
    require_file('./external_file_test.lua')
    return test()
end
