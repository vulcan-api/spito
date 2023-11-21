# SPITO
SPTIO (Simple Powerful Interprocess Tool to Overpower) has been designed
to allow users easily swap and edit configuration files.

## Use cases
**For a while it's a myth language**

### Simple config
- used rule: samba-guest-are-god
- purpose: use whole ready-to-use samba config
```shell
spito -r samba-guests-are-god
```

### Variable config
- used rule: wireguard
- purpose: use config that may need additional information
- good to notice: rule will generate keys for you

```shell
spito wireguard -o NUMBER_OF_CLIENT=3
```

### Part of config
- used rule: samba-guests-are-god
- purpose: use create mask from config

```shell
spito samba-guests-are-god@create-mask
```
## License
This project is licensed under the [GPL v3.0](./LICENSE)