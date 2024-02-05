# Spito

**This project is in pre-alpha and most of
the features described below does not work!**

Spito has been designed to allow users easily swap
and edit configuration files. It's user-friendly and
lays the foundations for auto-repair tools.

## Installation

For the whole operating system **(remember to run as root)**

```bash
export GOBIN=/usr/bin
go install
```

Only for local user

```bash
go install
export PATH=$PATH:~/go/bin
```

You can also add last line to .bashrc or other similar file.

## Use cases

### Simple config - samba

Samba is a simple FOSS file server. Samba is this kind of software
that requires configuration on start, and it costs a lot of searching
when you want to do anything.

**samba-guests-are-god:**
This rule allows samba guests to do almost everything. It's especially
useful when using virtual machines without shared files support.

Notice there is used official repository of samba rules stored on
[GitHub](https://github.com/avorty/spito-ruleset/tree/main/samba).

```shell
spito samba guests-are-god
```

### Variable config - wireguard

Wireguard is fast, modern, secure VPN tunnel. It's not hard to
configure, but requires some effort. Rule can generate required
private and public keys. Wireguard allows to easily
add clients, but rule add initially as many clients as
you specify.

Notice that johnny's repository of rules is stored on GitHub.

```shell
spito johnny/spitorules wireguard -o NUMBER_OF_CLIENT=3
```

### Part of config - gdm

Gnome is huge project that contains desktop environment,
display manager and default some application. Let's say
you want to install whole Gnome environment. But you can
also install and configure only gdm.

```shell
spito gitlab.com/johnny/dotfiles gnome@gdm
```

## License

This project is licensed under the [GPL v3.0](./LICENSE)
