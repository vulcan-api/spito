---
sidebar_position: 3
---

# api.fs

The `api.fs` module provides functions for working with the file system.

## api.fs.pathExists

### Arguments:
- `path` (string): The path to check.

### Returns:
- `exists` (bool): Whether the path exists.

### Example usage:

```lua
local exists = api.fs.pathExists("/etc/passwd")
```

## api.fs.fileExists

### Arguments:
- `path` (string): The path to check.
- `isDirectory` (bool): Whether the path is a directory.

### Returns:
- `exists` (bool): Whether the file exists.

### Example usage:

```lua
local exists = api.fs.fileExists("/etc/passwd", false)
```

## api.fs.readFile

### Arguments:
- `path` (string): The path to read.

### Returns:
- `content` (string): The content of the file.
- `error` (string): The error message if the file does not exist.

### Example usage:

```lua
function readPasswd()
  local content, err = api.fs.readFile("/etc/passwd")
  if err ~= nil then
    api.info.error("Error occured during reading the file: " .. err)
    return false
  end
  return true
end
```

## api.fs.readDir

### Arguments:
- `path` (string): The path to read.

### Returns:
- `files` ([]string): The files in the directory.
- `error` (string): The error message if the directory does not exist.

### Example usage:

```lua
function readDir()
  local files = api.fs.readDir("/etc")
  for _, file in ipairs(files) do
    api.info.info(file)
  end
end
```

## api.fs.fileContains

### Arguments:
- `fileContent` (string): The content of the file.
- `content` (string): The content to check.

### Returns:
- `contains` (bool): Whether the file contains the content.

### Example usage:

```lua
local contains = api.fs.fileContains(api.fs.readFile("/etc/passwd"), "root")
```

## api.fs.removeComments

### Arguments:
- `content` (string): The content to remove comments from.
- `singleLineStart` (string): The start of a single line comment.
- `multiLineStart` (string): The start of a multi line comment.
- `multiLineEnd` (string): The end of a multi line comment.

### Returns:
- `content` (string): The content without comments.

### Example usage:

```lua
local content = api.fs.removeComments(api.fs.readFile("/etc/passwd"), "#", "/*", "*/")
```

## api.fs.find

### Arguments:
- `regex` (string): The regex to search for.
- `fileContent` (string): The content to search in.

### Returns:
- `lines` ([]int): The lines where the regex was found.
- `error` (string): The error message if the regex is invalid.

### Example usage:

```lua
function findRoot()
  local lines, err = api.fs.find("root", api.fs.readFile("/etc/passwd"))
  if err ~= nil then
    api.info.error("Error occured during finding the regex: " .. err)
    return false
  end
  return true
end
```

## api.fs.findAll

### Arguments:
- `regex` (string): The regex to search for.
- `fileContent` (string): The content to search in.

### Returns:
- `lines` ([][]int): The lines where the regex was found.
- `error` - The error message if the regex is invalid.

### Example usage:

```lua
function findAllRoots()
  local lines, err = api.fs.findAll("root", api.fs.readFile("/etc/passwd"))
  if err ~= nil then
    api.info.error("Error occured during finding the regex: " .. err)
    return false
  end
  return true
end
```

## api.fs.getProperLines

### Arguments:
- `regex` (string): The regex to search for.
- `fileContent` (string): The content to search in.

### Returns:
- `lines` ([]string): The lines where the regex was found.
- `error` (string): The error message if the regex is invalid.

### Example usage:

```lua
function getRoots()
  local lines, err = api.fs.getProperLines("root", api.fs.readFile("/etc/passwd"))
  if err ~= nil then
    api.info.error("Error occured during finding the regex: " .. err)
    return false
  end
  return true
end
```

## api.fs.createFile

### Arguments:
- `path` (string): The path to create.
- `content` (string): The content of the file.
- `options` (CreateFileOptions): The options for creating the file.
- `error` (string): The error message if the file already exists.

`CreateFileOptions`:
- `optional` (bool): Whether the file is optional.
- `fileType` (string): The type of the file.

### Example usage:

```lua
function createFile()
  local err = api.fs.createFile("/etc/passwd", "root:x:0:0:root:/root:/bin/bash", { optional = false, fileType = "passwd" })
  if err ~= nil then
    api.info.error("Error occured during creating the file: " .. err)
    return false
  end
  return true
end
```

