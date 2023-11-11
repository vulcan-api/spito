# SPITO RULES specyfikacyja (API)
- getDistro() -> Struct (name and version)
- getPackageInfo() -> Struct (name, version, etc.)
- daemonInfo(string daemonName) -> Struct
- getInitSystem() -> string containing one of the following:
  + systemd
  + runit
  + openrc
  or empty string in case of errors

# fs
- read-only fs methods:
    - open()
    - close()
    - readFile()
