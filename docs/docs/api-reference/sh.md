---
sidebar_position: 5
---

# api.sh

The `api.sh` module provides functions for executing shell commands.

:::warning
This module works only if the rule is unsafe.
:::

## api.sh.command

### Arguments:
- `command` (string): The command to execute.

### Returns:
- `output` (string): The output of the command.
- `error` (string): The error message if the command fails.

### Example usage:

```lua
#![unsafe]

function ls()
  local output, err = api.sh.command("ls -l")
  if err ~= nil then
    api.info.Error("Error occured while executing the command: " .. err)
    return false
  end
  return true
end
```
