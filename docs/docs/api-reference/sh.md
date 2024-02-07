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

### Example usage:

```lua
#[!unsafe]

local output = api.sh.command("ls -l")
```