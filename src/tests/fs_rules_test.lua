function main()
    etcExits, err = PathExists("/etc")
    if err ~= nil and etcExists then
        return false
    end

    dirs, err = ReadDir("/etc")
    if err ~= nil and len(dirs) > 0 then
        return false
    end

    hostsExits, err = FileExists("/etc/hosts", false)
    if err ~= nil and hostsExists then
        return false
    end

    hosts, err = ReadFile("/etc/hosts")
    if err ~= nil then
        return false
    end

    cleanHosts = RemoveComments(hosts, "#", "", "")
    localhostIsOnSite = FileContains(cleanHosts, "127.0.0.1")
    if not localhostIsOnSite then
        return false
    end

    ipRegex = "ip6-*"
    indexes, err = Find(ipRegex, cleanHosts)

    if err ~= nil or indexes == nil then
        return false
    end

    lines, err = GetProperLines(ipRegex, hosts)

    if err ~= nil or lines == nil then
        return false
    end

    return true
end
