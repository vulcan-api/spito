function main()
    etcExits, err = api.fs.PathExists("/etc")
    if err ~= nil and etcExists then
        return false
    end

    dirs, err = api.fs.ReadDir("/etc")
    if err ~= nil and len(dirs) > 0 then
        return false
    end

    releaseExits, err = api.fs.FileExists("/etc/os-release", false)
    if err ~= nil and releaseExists then
        return false
    end

    resolv, err = api.fs.ReadFile("/etc/resolv.conf")
    if err ~= nil then
        return false
    end

    cleanResolv = api.fs.RemoveComments(resolv, "#", "", "")
    anyNameserverIsOnSite = api.fs.FileContains(cleanResolv, "nameserver")
    if not anyNameserverIsOnSite then
        return false
    end

    partitionRegex = "[0-9]+."
    indexes, err = api.fs.Find(partitionRegex, cleanResolv)

    if err ~= nil or indexes == nil then
        return false
    end

    lines, err = api.fs.GetProperLines(partitionRegex, resolv)

    if err ~= nil or lines == nil then
        return false
    end

    return true
end
