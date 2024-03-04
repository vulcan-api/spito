---
sidebar_position: 6
---

# api.daemon

The `api.daemon` module provides functions for working with daemons.

## api.daemon.get

### Arguments:
- `name` (string): The name of the daemon to get.

### Returns:
- `daemon` (Daemon): The daemon info.

### Example usage:

```lua
daemon, err = api.daemon.get("dbus")
if err then
    api.info.error("Error occured during obtaining daemon info!")
end
```

## api.daemon.start

### Arguments:
- `name` (string): The name of the daemon.

### Returns:
- `error` (string): The error message if the daemon could not be started.

### Example usage:

```lua
local err = api.daemon.start("httpd")
if err then
  api.info.error("Error occured during starting the daemon: " .. err)
end
```

## api.daemon.stop

### Arguments:
- `name` (string): The name of the daemon.

### Returns:
- `error` (string): The error message if the daemon could not be stopped.

### Example usage:

```lua
local err = api.daemon.stop("httpd")
if err then
  api.info.error("Error occured during stopping the daemon: " .. err)
end
```

## api.daemon.restart

### Arguments:
- `name` (string): The name of the daemon.

### Returns:
- `error` (string): The error message if the daemon could not be restarted.

### Example usage:

```lua
local err = api.daemon.restart("httpd")
if err then
  api.info.error("Error occured during restarting the daemon: " .. err)
end
```

## api.daemon.enable

### Arguments:
- `name` (string): The name of the daemon.

### Returns:
- `error` (string): The error message if the daemon could not be enabled.

### Example usage:

```lua
local err = api.daemon.enable("httpd")
if err then
  api.info.error("Error occured during enabling the daemon: " .. err)
end
```

## api.daemon.disable

### Arguments:
- `name` (string): The name of the daemon.

### Returns:
- `error` (string): The error message if the daemon could not be disabled.

### Example usage:

```lua
local err = api.daemon.disable("httpd")
if err then
  api.info.error("Error occured during disabling the daemon: " .. err)
end
```
