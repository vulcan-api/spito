function main()
    require_remote_result = require_remote("https://github.com/avorty/spito-ruleset", "dbus")
    require_result = require_file('./daemon_test.lua')
    return require_remote_result and require_result
end
