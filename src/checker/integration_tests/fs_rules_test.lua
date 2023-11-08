function main()
    etcExits, err = PathExists("/etc")
    if not err == nil and etcExists then
        return false
    end

    dirs, err = ReadDir("/etc")
    if not err == nil and len(dirs) > 0 then
        return false
    end

    hostsExits, err = FileExists("/etc/hosts", false)
    if not err == nil and hostsExists then
        return false
    end

    hosts, err = ReadFile("/etc/hosts")
    if not err == nil then
        return false
    end

    clearFile = RemoveComments(hosts, "#", "", "")
    localhostIsOnSite = FileContains(clearFile, "127.0.0.1")
    return localhostIsOnSite
end
