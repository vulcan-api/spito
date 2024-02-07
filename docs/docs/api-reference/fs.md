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

### Example usage:

```lua
local content = api.fs.readFile("/etc/passwd")
```

## api.fs.readDir

### Arguments:
- `path` (string): The path to read.

### Returns:
- `files` ([]string): The files in the directory.

### Example usage:

```lua
local files = api.fs.readDir("/etc")
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

### Example usage:

```lua
local lines = api.fs.find("root", api.fs.readFile("/etc/passwd"))
```

## api.fs.findAll

### Arguments:
- `regex` (string): The regex to search for.
- `fileContent` (string): The content to search in.

### Returns:
- `lines` ([][]int): The lines where the regex was found.

### Example usage:

```lua
local lines = api.fs.findAll("root", api.fs.readFile("/etc/passwd"))
```

## api.fs.getProperLines

### Arguments:
- `regex` (string): The regex to search for.
- `fileContent` (string): The content to search in.

### Returns:
- `lines` ([]string): The lines where the regex was found.

### Example usage:

```lua
local lines = api.fs.getProperLines("root", api.fs.readFile("/etc/passwd"))
```

## api.fs.createFile

### Arguments:
- `path` (string): The path to create.
- `content` (string): The content of the file.
- `options` (CreateFileOptions): The options for creating the file.

`CreateFileOptions`:
- `optional` (bool): Whether the file is optional.
- `fileType` (string): The type of the file.

### Example usage:

```lua
api.fs.createFile("/etc/passwd", "root:x:0:0:root:/root:/bin/bash", { optional = false, fileType = "passwd" })
```

