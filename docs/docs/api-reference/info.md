---
sidebar_position: 4
---

# api.info

The `api.info` module provides functions for logging information.

When executing from CLI the log messages are printed to the console. <br />
When executing from the GUI the log messages are displayed in the app. 

## log

### Arguments:
- `messages` (string...): The message to log.

### Example usage:

```lua
api.info.log("Hello, world!")
```

## debug

### Arguments:
- `messages` (string...): The message to log.

### Example usage:

```lua
api.info.debug("Hello, world!")
```

## error

### Arguments:
- `messages` (string...): The message to log.

### Example usage:

```lua
api.info.error("Hello, world!")
```

## warn

### Arguments:
- `messages` (string...): The message to log.

### Example usage:

```lua
api.info.warn("Hello, world!")
```

## important

### Arguments:
- `messages` (string...): The message to log.

### Example usage:

```lua
api.info.important("Hello, world!")
```






