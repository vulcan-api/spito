#![Unsafe]

function main()
    local destPath = "/tmp/spito-test/nfdsa321980"
    local err = api.git.clone("https://github.com/Avorty/.github", destPath)
    if err then
        api.info.error(err)
        return false
    end

    local foundReadme = false

    local files, err = api.fs.readDir("/tmp/spito-test/nfdsa321980")
    if err then
        api.info.error(err)
        return false
    end

    for _, file in files() do
        if string.find(file.Path, "README.md") then
            foundReadme = true
        end
    end

    local _, err = api.sh.command("rm -rf " .. destPath)
    if err then
        api.info.error(err)
        return false
    end

    return foundReadme
end
