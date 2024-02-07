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

### Example usage:

```lua
local daemon = api.sys.getDaemon("NetworkManager")
```

## api.sys.getInitSystem

### Arguments:
- `name` (string): The name of the package to get.

### Returns:
- `initSystem` (InitSystem): The init system info.

### Example usage:

```lua
local initSystem = api.sys.getInitSystem()
```

