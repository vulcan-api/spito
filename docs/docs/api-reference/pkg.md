---
sidebar_position: 1
---

# api.pkg

The `api.pkg` module provides functions for working with packages.

## get

### Arguments:
- `name` (string): The name of the package to get.

### Returns:
- `package` (Package): The package info from `pacman -Qi`.
- `error` (error): The error message if the package does not exist.


| Field         | Type    | Description |
|---------------|---------|-------------|
| Name          | string  | The name of the package. |
| Version       | string  | The version of the package. |
| Description   | string  | The description of the package. |
| Architecture  | string  | The architecture of the package. |
| URL           | string  | The URL of the package. |
| Licenses      | []string| The licenses of the package. |
| Groups        | []string| The groups of the package. |
| Provides      | []string| The packages provided by the package. |
| DependsOn     | []string| The packages the package depends on. |
| OptionalDeps  | []string| The optional dependencies of the package. |
| RequiredBy    | []string| The packages that require the package. |
| OptionalFor   | []string| The packages the package is optional for. |
| ConflictsWith | []string| The packages the package conflicts with. |
| Replaces      | []string| The packages the package replaces. |
| InstalledSize | []string| The installed size of the package. |
| Packager      | string  | The packager of the package. |
| BuildDate     | string  | The build date of the package. |
| InstallDate   | string  | The install date of the package. |
| InstallReason | string  | The install reason of the package. |
| InstallScript | bool    | Whether the package has an install script. |
| ValidatedBy   | string  | The signature of the package. |

### Example usage:

```lua
function gnome_exists()
  app_info, err = api.pkg.Get("gnome-shell")
  if err ~= nil then
    api.info.Error("Error occured during obtaining package info!")
    return false
  end
  return true
end
```
