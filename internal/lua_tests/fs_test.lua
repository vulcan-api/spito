function main()
    etcExits, err = api.fs.PathExists("/etc")
    if err ~= nil and etcExists then
        api.info.Error(err)
        return false
    end

    dirs, err = api.fs.ReadDir("/etc")
    if err ~= nil and len(dirs) > 0 then
        api.info.Error(err)
        return false
    end

    releaseExits, err = api.fs.FileExists("/etc/os-release", false)
    if err ~= nil and releaseExists then
        api.info.Error(err)
        return false
    end

    resolv, err = api.fs.ReadFile("/etc/resolv.conf")
    if err ~= nil then
        api.info.Error(err)
        return false
    end

    cleanResolv = api.fs.RemoveComments(resolv, "#", "", "")
    anyNameserverIsOnSite = api.fs.FileContains(cleanResolv, "nameserver")
    if not anyNameserverIsOnSite then
        api.info.Error("No nameserver is on site")
        return false
    end

    partitionRegex = "[0-9]+."
    indexes, err = api.fs.Find(partitionRegex, cleanResolv)

    if err ~= nil or indexes == nil then
        api.info.Error(err)
        return false
    end

    lines, err = api.fs.GetProperLines(partitionRegex, resolv)

    if err ~= nil or lines == nil then
        api.info.Error(err)
        return false
    end

    configPath = "/tmp/spito-lua-test/example.json"
    options = {
        ConfigType = api.fs.Config.Json
    }

    err = api.fs.UpdateConfig(configPath, '{"example-key":"example-val"}', options)
    if err ~= nil then
        api.info.Error(err)
        return false
    end

    err = api.fs.CreateConfig(configPath, '{"example-key":"example-val"}', options)
    if err ~= nil then
        api.info.Error(err)
        return false
    end

    err = api.fs.CreateConfig(configPath, '{"next-example-key":"next-example-val"}', options)
    if err ~= nil then
        api.info.Error(err)
        return false
    end

    content, err = api.fs.ReadFile(configPath)
    if err ~= nil then
        api.info.Error(err)
        return false
    end

    err = api.fs.CompareConfigs(content, '{"example-key": "example-val", "next-example-key": "next-example-val", "first-key":"first-val"}', options.ConfigType)
    if err ~= nil then
        api.info.Error(err)
        return false
    end

    return true
end
