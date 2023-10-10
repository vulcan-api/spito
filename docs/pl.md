# SPITO RULES specyfikacyja (API)
- getCurrentDistro() -> Struct (name and version)
    - zwraca jakie mamy distro
- isPackageInstalled(string version) -> bool
    - ten grzesio powinien sprawdzać lokalnym, natywnym package managerem, czy dany pakiet jest zainstalowany 
    - distro-dependent
- isServiceRunning() -> bool # idk czy to ma sens (daemonInfo imo better)
    - ten grzybogniew należy, aby sprawdził, czy chodzi jakiś service
    - distro-dependent (ale i tak głównie używają ludzie systemd)
- daemonInfo(string daemonName) -> Struct
    - Zwraca możliwie jak najwięcej info o daemonie
  # fs
- read-only api
