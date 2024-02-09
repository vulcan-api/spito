---
sidebar_position: 2
---

# api.sys

The `api.sys` module provides functions for working with the system.

## .getDistro

### Arguments:
- `name` (string): The name of the package to get.

### Returns:
- `distro` (Distro): The distro info.

### Example usage:

```lua
distro = api.sys.getDistro()
```

## .getDaeomon

### Arguments:
- `name` (string): The name of the package to get.

### Returns:
- `daemon` (Daemon): The daemon info.
- `error` (error): The error message if the daemon does not exist.

### Example usage:

```lua
function networkManagerExists()
  daemon, err = api.sys.getDaemon("dbus")
  if err ~= nil then
    api.info.error("Error occured during obtaining daemon info!")
    return false
  end
  return true
end
```

## .getInitSystem

### Arguments:
- `name` (string): The name of the package to get.

### Returns:
- `initSystem` (InitSystem): The init system info.
- `error` (error): The error message if the init system does not exist.

### Example usage:

```lua
function initSystemExists()
  initSystem, err = api.sys.getInitSystem()
  if err ~= nil then
    api.info.error("Error occured during obtaining init system info!")
    return false
  end
  return true
end
```
