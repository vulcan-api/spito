---
sidebar_position: 2
---

# api.sys

The `api.sys` module provides functions for working with the system.

## api.sys.getDistro

### Arguments:
- `name` (string): The name of the package to get.

### Returns:
- `distro` (Distro): The distro info.

### Example usage:

```lua
local distro = api.sys.getDistro()
```

## api.sys.getDaeomon

### Arguments:
- `name` (string): The name of the package to get.

### Returns:
- `daemon` (Daemon): The daemon info.
- `error` (string): The error message if the daemon does not exist.

### Example usage:

```lua
function networkManagerExists()
  local daemon, err = api.sys.getDaemon("dbus")
  if err ~= nil then
    api.info.error("Error occured during obtaining daemon info!")
    return false
  end
  return true
end
```

## api.sys.getInitSystem

### Arguments:
- `name` (string): The name of the package to get.

### Returns:
- `initSystem` (InitSystem): The init system info.
- `error` (string): The error message if the init system does not exist.

### Example usage:

```lua
function initSystemExists()
  local initSystem, err = api.sys.getInitSystem()
  if err ~= nil then
    api.info.error("Error occured during obtaining init system info!")
    return false
  end
  return true
end
```
