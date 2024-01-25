function main()
    local dbus = api.sys.GetDaemon("dbus")
    if dbus.Name == "" or not dbus.IsActive or not dbus.IsEnabled then
        return false
    end
    return true
end
