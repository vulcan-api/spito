function main()
    etcExits, err = api.fs.pathExists("/etc")
    if err ~= nil and etcExists then
        api.info.error(err)
        return false
    end

    dirs, err = api.fs.readDir("/etc")
    if err ~= nil and len(dirs) > 0 then
        api.info.error(err)
        return false
    end

    releaseExits, err = api.fs.fileExists("/etc/os-release", false)
    if err ~= nil and releaseExists then
        api.info.error(err)
        return false
    end

    resolv, err = api.fs.readFile("/etc/resolv.conf")
    if err ~= nil then
        api.info.error(err)
        return false
    end

    cleanResolv = api.fs.removeComments(resolv, "#", "", "")
    anyNameserverIsOnSite = api.fs.fileContains(cleanResolv, "nameserver")
    if not anyNameserverIsOnSite then
        api.info.error("No nameserver is on site")
        return false
    end

    partitionRegex = "[0-9]+."
    indexes, err = api.fs.find(partitionRegex, cleanResolv)

    if err ~= nil or indexes == nil then
        api.info.error(err)
        return false
    end

    lines, err = api.fs.getProperLines(partitionRegex, resolv)
    if err ~= nil or lines == nil then
        api.info.error(err)
        return false
    end

    fileToBeCreatedPath = "/tmp/spito-create-file-test-24tc89t221"
    fileToBeCreatedContent = "example content"
    err = api.fs.createFile(fileToBeCreatedPath, fileToBeCreatedContent, false)
    if err ~= nil then
        api.info.error(err)
        return false
    end

    content, err = api.fs.readFile(fileToBeCreatedPath)
    if err ~= nil then
        api.info.error(err)
        return false
    end
    if content ~= fileToBeCreatedContent then
        api.info.error("Failed to property create file - wrong content")
        return false
    end


    configPath = "/tmp/spito-lua-test/example.json"
    options = {
        ConfigType = api.fs.config.json
    }

    err = api.fs.updateConfig(configPath, '{"example-key":"example-val"}', options)
    if err ~= nil then
        api.info.error(err)
        return false
    end

    err = api.fs.createConfig(configPath, '{"example-key":"example-val"}', options)
    if err ~= nil then
        api.info.error(err)
        return false
    end

    err = api.fs.createConfig(configPath, '{"next-example-key":"next-example-val"}', options)
    if err ~= nil then
        api.info.error(err)
        return false
    end

    content, err = api.fs.readFile(configPath)
    if err ~= nil then
        api.info.error(err)
        return false
    end

    err = api.fs.compareConfigs(content,
        '{"example-key": "example-val", "next-example-key": "next-example-val", "first-key":"first-val"}',
        options.ConfigType)
    if err ~= nil then
        api.info.error(err)
        return false
    end

    return true
end
