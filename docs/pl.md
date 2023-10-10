# SPITO RULES specyfikacyja (API)
- getCurrentDistro() -> Struct (name and version)
    - zwraca jakie mamy distro
- isPackageInstalled(string version) -> bool
    - ten grzesio powinien sprawdzać lokalnym, natywnym package managerem, czy dany pakiet jest zainstalowany 
    - distro-dependent
- daemonInfo(string daemonName) -> Struct
    - Zwraca możliwie jak najwięcej info o daemonie
  # fs
- read-only api
